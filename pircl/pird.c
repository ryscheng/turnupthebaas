#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <time.h>
#include <unistd.h>
#include <sys/mman.h>
#include <sys/types.h>
#include <sys/shm.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/un.h>
#ifdef __APPLE__
#include <OpenCL/opencl.h>
#else
#include <CL/cl.h>
#endif

#define DATA_TYPE unsigned long

#define SOCKET_NAME "pir.socket"

////////////////////////////////////////////////////////////////////////////////
//
// PIR Server.
//
// Communication occurs via a socket. several commands are present.
// 1. read. expects an input pir vector.
// 2. configure. sets parameters for batching of groups.
// 3. write. updates the database.
////////////////////////////////////////////////////////////////////////////////
int cell_length;
int cell_count;
int batch_size;
size_t workgroup_size;
char* invector;
DATA_TYPE* database;
DATA_TYPE* output;
cl_context context;                 // compute context
cl_command_queue commands;          // compute command queue
cl_program program;                 // compute program
cl_kernel kernel;                   // compute kernel
cl_mem gpu_db;                      // device memory used for the database
cl_mem gpu_input;
cl_mem gpu_output;

int configure(int, int, int);
int do_write(DATA_TYPE*);
DATA_TYPE* do_read(char*);

int main()
{
    int err;

    int socket_fd = socket(AF_UNIX, SOCK_STREAM, 0);
    int client_sock;
    if (socket_fd == -1) {
      printf("Error: Failed to create sockets!\n");
      return EXIT_FAILURE;
    }

    // Bind the socket for communication.
    unlink(SOCKET_NAME);
    struct sockaddr_un socket_name;
    memset(&socket_name, 0, sizeof(struct sockaddr_un));
    socket_name.sun_family = AF_UNIX;
    strncpy(socket_name.sun_path, SOCKET_NAME, sizeof(socket_name.sun_path) - 1);
    err = bind(socket_fd, (const struct sockaddr *) &socket_name, sizeof(struct sockaddr_un));
    if (err == -1) {
      printf("Error: Failed to bind socket!\n");
      return EXIT_FAILURE;
    }

    err = listen(socket_fd, 1);
    if (err == -1) {
      printf("Error: Failed to listen on socket!\n");
      return EXIT_FAILURE;
    }


    char next_command;
    int ret;
    int dbhndl;
    int configuration_params[3];
    for (;;) {
      client_sock = accept(socket_fd, NULL, NULL);
      if (client_sock == -1) {
        printf("Error: failed to accept client on listening socket!\n");
        return EXIT_FAILURE;
      }
      for(;;) {
        /* Wait for next data packet. */
        ret = read(client_sock, &next_command, 1);
        if (ret == -1 || ret == 0) {
          printf("Client disconnected.\n");
          break;
        }
        if (next_command == '1') { //read
          ret = read(client_sock, invector, cell_count * batch_size);
          if (ret == -1) {
            printf("Failed to read configuration.\n");
            break;
          }
          output = do_read(invector);
          ret = write(client_sock, output, cell_length * batch_size);
          if (ret == -1) {
            printf("Failed to write response.\n");
            break;
          }
        } else if (next_command == '2') { //configure
          ret = read(client_sock, configuration_params, 3*sizeof(int));
          if (ret == -1) {
            printf("Failed to read configuration.\n");
            break;
          }
          configure(configuration_params[0], configuration_params[1], configuration_params[2]);
        } else if (next_command == '3') { // write
          if (dbhndl != 0) {
            shmdt(database);
          }
          ret = read(client_sock, &dbhndl, sizeof(dbhndl));
          if (ret == -1) {
            printf("Failed to learn shm id.\n");
            break;
          }
          database = shmat(dbhndl, NULL, SHM_RDONLY);
          if (database == -1) {
            printf("Failed to open shm ptr: %d.\n", errno);
            break;
          }
          do_write(database);
        } else {
          printf("Unexpected command: %c\n", next_command);
          break;
        }
      }
    }
}

int configure(int n_cell_length, int n_cell_count, int n_batch_size) {
  int err;

  cell_length = n_cell_length;
  cell_count = n_cell_count;
  batch_size = n_batch_size;

  if (output != NULL) {
    free(output);
    free(invector);
    invector = NULL;
    output = NULL;
  }
  output = malloc(cell_length * batch_size);
  invector = malloc(cell_count * batch_size);

  if (context != NULL) {
    clReleaseProgram(program);
    clReleaseKernel(kernel);
    clReleaseCommandQueue(commands);
    clReleaseContext(context);
  }

  cl_device_id device_id;             // compute device id
  cl_platform_id cl_platform;

  // Connect to a compute device
  err = clGetPlatformIDs(1, &cl_platform, NULL);
  int gpu = 1;
  err = clGetDeviceIDs(cl_platform, gpu ? CL_DEVICE_TYPE_GPU : CL_DEVICE_TYPE_CPU, 1, &device_id, NULL);
  if (err != CL_SUCCESS)
  {
      printf("Error: Failed to create a device group!\n");
      return EXIT_FAILURE;
  }

  // Create a compute context
  context = clCreateContext(0, 1, &device_id, NULL, NULL, &err);
  if (!context)
  {
      printf("Error: Failed to create a compute context!\n");
      return EXIT_FAILURE;
  }

  // Create a command queue
  commands = clCreateCommandQueue(context, device_id, 0, &err);
  if (!commands)
  {
      printf("Error: Failed to create a command commands!\n");
      return EXIT_FAILURE;
  }

  // Map in kernel source code
  struct stat kernelStat;
  int kernelHandle = open("pir_kernel.c", O_RDONLY);
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
  //err = clGetDeviceInfo(device_id, CL_DEVICE_MAX_MEM_ALLOC_SIZE, sizeof(cl_ulong), &buf_max_size, NULL);
  //err = clGetDeviceInfo(device_id, CL_DEVICE_GLOBAL_MEM_SIZE, sizeof(cl_ulong), &dev_mem_size, NULL);
  //printf("Global size is %lu. Maximum buffer on device can be %lu.\n", dev_mem_size, buf_max_size);

  // Get the maximum work group size for executing the kernel on the device
  //
  err = clGetKernelWorkGroupInfo(kernel, device_id, CL_KERNEL_WORK_GROUP_SIZE, sizeof(workgroup_size), &workgroup_size, NULL);
  if (err != CL_SUCCESS)
  {
      printf("Error: Failed to retrieve kernel work group info! %d\n", err);
      exit(1);
  }

  if (gpu_db != NULL) {
    clReleaseMemObject(gpu_db);
    clReleaseMemObject(gpu_input);
    clReleaseMemObject(gpu_output);
  }
  gpu_db = clCreateBuffer(context, CL_MEM_READ_ONLY, cell_length * cell_count, NULL, NULL);
  gpu_input = clCreateBuffer(context,  CL_MEM_READ_ONLY, cell_count * batch_size, NULL, NULL);
  gpu_output = clCreateBuffer(context, CL_MEM_WRITE_ONLY, cell_length * batch_size, NULL, NULL);

  if (!gpu_db || !gpu_input || !gpu_output) {
      printf("Error: Failed to allocate device memory!\n");
      exit(1);
  }

  // set kernel args.
  unsigned int db_cnt = cell_count * cell_length / sizeof(DATA_TYPE);
  unsigned int db_cell_cnt = cell_length / sizeof(DATA_TYPE);
  err = 0;
  err  = clSetKernelArg(kernel, 0, sizeof(cl_mem), &gpu_db);
  err |= clSetKernelArg(kernel, 1, sizeof(cl_mem), &gpu_input);
  err |= clSetKernelArg(kernel, 2, cell_length, NULL);
  err |= clSetKernelArg(kernel, 3, sizeof(unsigned int), &db_cnt);
  err |= clSetKernelArg(kernel, 4, sizeof(unsigned int), &db_cell_cnt);
  err |= clSetKernelArg(kernel, 5, sizeof(cl_mem), &gpu_output);
  if (err != CL_SUCCESS)
  {
      printf("Error: Failed to set kernel arguments! %d\n", err);
      exit(1);
  }
  printf("Reconfigured. Database now %d items of %d bytes. Batches of %d requests.\n", cell_count, cell_length, batch_size);
}

int do_write(DATA_TYPE* db) {
  int err;
  err = clEnqueueWriteBuffer(commands, gpu_db, CL_TRUE, 0, cell_length * cell_count, db, 0, NULL, NULL);
  if (err != CL_SUCCESS) {
    printf("Error: Failed to write db to device!\n");
    exit(1);
  }
  printf("Database updated.\n");
}

DATA_TYPE* do_read(char* invector) {
  int err;

  // Write input vector.
  err = clEnqueueWriteBuffer(commands, gpu_input, CL_TRUE, 0, cell_count * batch_size, invector, 0, NULL, NULL);
  if (err != CL_SUCCESS)
  {
      printf("Error: Failed to write to source array!\n");
      exit(1);
  }

  // Execute the kernel over the entire range of our 1d input data set
  // using the maximum number of work group items for this device
  //
  size_t totaljobs = workgroup_size * batch_size;
  err = clEnqueueNDRangeKernel(commands, kernel, 1, NULL, &totaljobs, &workgroup_size, 0, NULL, NULL);
  if (err) {
      printf("Error: Failed to execute kernel!\n");
      exit(1);
  }

  // Read back the results from the device to verify the output
  //
  err = clEnqueueReadBuffer(commands, gpu_output, CL_TRUE, 0, cell_length * batch_size, output, 0, NULL, NULL );
  if (err != CL_SUCCESS)
  {
      printf("Error: Failed to read output array! %d\n", err);
      exit(1);
  }
  printf("Read Batch.\n");
  return output;
}
