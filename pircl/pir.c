#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/mman.h>
#ifdef __APPLE__
#include <OpenCL/opencl.h>
#else
#include <CL/cl.h>
#endif

#define BLOCK_COUNT (1024)
#define BLOCK_SIZE (1024)

#define DATA_TYPE unsigned long

////////////////////////////////////////////////////////////////////////////////
int main(int argc, char** argv)
{
    int err;                            // error code returned from api calls

    DATA_TYPE database[BLOCK_COUNT * BLOCK_SIZE];
    char invector[BLOCK_COUNT]; // input vector given to device
    DATA_TYPE outdata[BLOCK_SIZE];   // results returned from device
    unsigned int correct;               // number of correct results returned

    size_t global;                      // global domain size for our calculation
    size_t local;                       // local domain size for our calculation

    cl_device_id device_id;             // compute device id
    cl_context context;                 // compute context
    cl_command_queue commands;          // compute command queue
    cl_program program;                 // compute program
    cl_kernel kernel;                   // compute kernel

    cl_mem db;                          // device memory used for the database
    cl_mem input;                       // device memory used for the input array
    cl_mem output;                      // device memory used for the output array

    cl_ulong buf_max_size;
    cl_ulong dev_mem_size;

    // Fill our input vectors
    int i = 0;
    int j = 0;
    unsigned int count = BLOCK_COUNT;
    unsigned int dbsize = BLOCK_COUNT * BLOCK_SIZE;
    unsigned int block_size = BLOCK_SIZE;
    for(i = 0; i < count; i++)
        invector[i] = rand();

    // Connect to a compute device
    //
    int gpu = 1;
    err = clGetDeviceIDs(NULL, gpu ? CL_DEVICE_TYPE_GPU : CL_DEVICE_TYPE_CPU, 1, &device_id, NULL);
    if (err != CL_SUCCESS)
    {
        printf("Error: Failed to create a device group!\n");
        return EXIT_FAILURE;
    }

    // Create a compute context
    //
    context = clCreateContext(0, 1, &device_id, NULL, NULL, &err);
    if (!context)
    {
        printf("Error: Failed to create a compute context!\n");
        return EXIT_FAILURE;
    }

    // Create a command commands
    //
    commands = clCreateCommandQueue(context, device_id, 0, &err);
    if (!commands)
    {
        printf("Error: Failed to create a command commands!\n");
        return EXIT_FAILURE;
    }

    // Map in kernel source code
    struct stat kernelStat;
    int kernelHandle = open("kernel.c", O_RDONLY);
    int status = fstat (kernelHandle, &kernelStat);
    char* kernelSource = (char *) mmap(NULL, kernelStat.st_size, PROT_READ, MAP_PRIVATE, kernelHandle, 0);

    // Create the compute program from the source buffer
    //
    program = clCreateProgramWithSource(context, 1, (const char **) & kernelSource, NULL, &err);
    if (!program)
    {
        printf("Error: Failed to create compute program!\n");
        return EXIT_FAILURE;
    }

    // Build the program executable
    //
    err = clBuildProgram(program, 0, NULL, NULL, NULL, NULL);
    if (err != CL_SUCCESS)
    {
        size_t len;
        char buffer[2048];

        printf("Error: Failed to build program executable!\n");
        clGetProgramBuildInfo(program, device_id, CL_PROGRAM_BUILD_LOG, sizeof(buffer), buffer, &len);
        printf("%s\n", buffer);
        exit(1);
    }

    // Create the compute kernels in the program we wish to run
    //
    kernel = clCreateKernel(program, "pir", &err);
    if (!kernel || err != CL_SUCCESS)
    {
        printf("Error: Failed to create compute kernel!\n");
        exit(1);
    }

    // learn about the device memory size:
    err = clGetDeviceInfo(device_id, CL_DEVICE_MAX_MEM_ALLOC_SIZE, sizeof(cl_ulong), &buf_max_size, NULL);
    printf("Max Alloc size is %llu. Allocated size will be %d.\n", buf_max_size, BLOCK_COUNT * BLOCK_SIZE);
    err = clGetDeviceInfo(device_id, CL_DEVICE_GLOBAL_MEM_SIZE, sizeof(cl_ulong), &dev_mem_size, NULL);
    printf("Global size is %llu.\n", dev_mem_size);


    // Create the arrays in device memory for our calculation
    //
    db = clCreateBuffer(context, CL_MEM_READ_ONLY, sizeof(DATA_TYPE) * BLOCK_COUNT * BLOCK_SIZE, NULL, NULL);
    input = clCreateBuffer(context,  CL_MEM_READ_ONLY, count, NULL, NULL);
    output = clCreateBuffer(context, CL_MEM_WRITE_ONLY, sizeof(DATA_TYPE) * BLOCK_SIZE, NULL, NULL);
    if (!input || !output || !db)
    {
        printf("Error: Failed to allocate device memory!\n");
        exit(1);
    }

    // Initialize the database.
    printf("Allocating & randomizing memory for the database.\n");
    for(i = 0; i < dbsize; i++)
        database[i] = rand();

    err = clEnqueueWriteBuffer(commands, db, CL_TRUE, 0, dbsize, database, 0, NULL, NULL);
    if (err != CL_SUCCESS)
    {
        printf("Error: Failed to write db to device!\n");
        exit(1);
    }

    printf("Done.\n");

    // Write our data set into the input array in device memory
    err = clEnqueueWriteBuffer(commands, input, CL_TRUE, 0,  count, invector, 0, NULL, NULL);
    if (err != CL_SUCCESS)
    {
        printf("Error: Failed to write to source array!\n");
        exit(1);
    }

    // Set the arguments to our compute kernel
    //
    err = 0;
    err  = clSetKernelArg(kernel, 0, sizeof(cl_mem), &db);
    err |= clSetKernelArg(kernel, 1, sizeof(cl_mem), &input);
    err |= clSetKernelArg(kernel, 2, sizeof(unsigned long) * BLOCK_SIZE, NULL);
    err |= clSetKernelArg(kernel, 3, sizeof(unsigned int), &dbsize);
    err |= clSetKernelArg(kernel, 4, sizeof(unsigned int), &block_size);
    err |= clSetKernelArg(kernel, 5, sizeof(cl_mem), &output);
    if (err != CL_SUCCESS)
    {
        printf("Error: Failed to set kernel arguments! %d\n", err);
        exit(1);
    }

    // Get the maximum work group size for executing the kernel on the device
    //
    err = clGetKernelWorkGroupInfo(kernel, device_id, CL_KERNEL_WORK_GROUP_SIZE, sizeof(local), &local, NULL);
    if (err != CL_SUCCESS)
    {
        printf("Error: Failed to retrieve kernel work group info! %d\n", err);
        exit(1);
    }

    // Execute the kernel over the entire range of our 1d input data set
    // using the maximum number of work group items for this device
    //
    global = count;
    err = clEnqueueNDRangeKernel(commands, kernel, 1, NULL, &global, &local, 0, NULL, NULL);
    if (err)
    {
        printf("Error: Failed to execute kernel!\n");
        return EXIT_FAILURE;
    }
    // Wait for the command commands to get serviced before stage 2
    //
    clFinish(commands);

    // Read back the results from the device to verify the output
    //
    err = clEnqueueReadBuffer( commands, output, CL_TRUE, 0, sizeof(DATA_TYPE) * BLOCK_SIZE, outdata, 0, NULL, NULL );
    if (err != CL_SUCCESS)
    {
        printf("Error: Failed to read output array! %d\n", err);
        exit(1);
    }

    printf("Recomputing to make sure these look right.\n");

    // Validate our results
    //
    correct = 0;
    DATA_TYPE acc[BLOCK_SIZE];
    int bit;
    for(i = 0; i < 1; i++)
    {
        for (j = 0; j < BLOCK_SIZE; j++)
            acc[j] = 0;
        for (j = 0; j < dbsize; j++) {
          bit = j / BLOCK_SIZE;
          if (invector[bit] != 0)
              acc[j % BLOCK_SIZE] ^= database[j];
        }
        for (j = 0; j < BLOCK_SIZE; j++)
            if(outdata[j] != acc[j])
              goto OUTERCONTINUE;
        correct++;
        OUTERCONTINUE:
        ;
    }

    // Print a brief summary detailing the results
    //
    if (correct > 0)
      printf("Computed the correct value!\n");

    // Shutdown and cleanup
    //
    clReleaseMemObject(db);
    clReleaseMemObject(input);
    clReleaseMemObject(output);
    clReleaseProgram(program);
    clReleaseKernel(kernel);
    clReleaseCommandQueue(commands);
    clReleaseContext(context);

    return 0;
}
