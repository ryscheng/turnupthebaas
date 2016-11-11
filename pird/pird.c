#include <errno.h>
#include <fcntl.h>
#include <math.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/types.h>
#include <sys/shm.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/un.h>
#include <time.h>
#include <unistd.h>
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
//
// Two sets of input/output buffers are present as an attempt to keep the GPU
// busy at all times. This attempts to make use of the memory optimzations
// recommended by nvidia:
// * Frequent data transfer (read PIR vectors, and responding cells) will have
//   highest bandwidth when performed on pinned main memory pages chosen by the
//   gpu using CL_MEM_ALLOC_HOST_PTR.
// * New gpu's support overlapped transfer and device computation, suggesting
//   that input & output should occur asynchronously with kernel execution to
//   optimize throughput. Since this is evidenced by the ability to potentially
//   simultaneously copy in both directs and compute, we attempt to use this
//   capability in two ways:
//   * calls to clEnqueueReadBuffer and clEnqueueWriteBuffer are asynchronous to
//   put both transfers and computation in the command queue at the same time.
//   * there are two instances of the kernel with two parallel buffers, so that
//   the output of one can be read back to the host and input to the next
//   request copied to the device while the other is executing.
//   If the two parallel executions are logically:
//   read 1                  read 2
//   compute 1               compute 2
//   write 1                 write 2
//   We optimize throughput by interleaving as:
//   read 1
//   compute 1
//   read 2
//   (steady-state)
//   compute 2
//   read 1
//   write 1    <blocking>
//   compute 1
//   read 2
//   write 2    <blocking>
//   (end of steady-state)
//
//   The point here is to push the retreival of the write as late as possible
//   so that work for the other execution instance is able to complete while
//   it is happening. In doing so, it's okay for it to block, because we
//   wouldn't be able to start the next computation until it's done anyway
//   since doing so would overwrite the device buffer we're copying out.
////////////////////////////////////////////////////////////////////////////////

int cell_length;
int cell_count;
int batch_size;
size_t workgroup_size;

typedef enum { FALSE, TRUE } bool;

typedef struct {
  bool input_loaded;
  bool output_dirty;
  unsigned char* input;
  DATA_TYPE* output;
  cl_mem input_pin;
  cl_mem output_pin;
  cl_mem gpu_input;
  cl_mem gpu_output;
  cl_kernel kernel;
} pipeline_t;

DATA_TYPE* database;
cl_context context;                 // compute context
cl_command_queue commands;          // compute command queue
cl_program program;                 // compute program
cl_mem gpu_db;                      // device memory used for the database
pipeline_t* pipelines;

int configure(int, int, int, int, int);
int do_write(DATA_TYPE*);
void listDevices();

static volatile char* socketpath;
void intHandler(int dummy) {
  unlink((char*)socketpath);
  exit(1);
}

// Set up a pipeline (read and write buffers + compute kernel.)
// Prerequisite: the 'program' should have been compiled by TODO.
int pipeline_init(pipeline_t* pipeline) {
  int err;
  pipeline->kernel = clCreateKernel(program, "pir", &err);
  if (!pipeline->kernel || err != CL_SUCCESS) {
    printf("Error: Failed to create compute kernel!\n");
    return -1;
  }

  pipeline->input_pin = clCreateBuffer(context,
                                      CL_MEM_READ_ONLY | CL_MEM_ALLOC_HOST_PTR,
                                      cell_count * batch_size / 8,
                                      NULL,
                                      NULL);
  pipeline->gpu_input = clCreateBuffer(context,
                                      CL_MEM_READ_ONLY,
                                      cell_count * batch_size / 8,
                                      NULL,
                                      NULL);

  pipeline->output_pin = clCreateBuffer(context,
                                       CL_MEM_WRITE_ONLY | CL_MEM_ALLOC_HOST_PTR,
                                       cell_length * batch_size,
                                       NULL,
                                       NULL);
  pipeline->gpu_output = clCreateBuffer(context,
                                       CL_MEM_WRITE_ONLY,
                                       cell_length * batch_size,
                                       NULL,
                                       NULL);
  if (!pipeline->input_pin || !pipeline->output_pin ||
      !pipeline->gpu_input || !pipeline->gpu_output) {
    printf("Error: Failed to create pipeline buffers.\n");
    return -2;
  }

  pipeline->input = (unsigned char *) clEnqueueMapBuffer(commands,
      pipeline->input_pin,
      CL_TRUE,
      CL_MAP_WRITE,
      0,
      cell_count * batch_size / 8,
      0,
      NULL,
      NULL,
      NULL);
  pipeline->output = (DATA_TYPE*) clEnqueueMapBuffer(commands,
      pipeline->output_pin,
      CL_TRUE,
      CL_MAP_READ,
      0,
      cell_length * batch_size,
      0,
      NULL,
      NULL,
      NULL);
  if (!pipeline->input || !pipeline->output) {
    printf("Error: Failed to map pipeline buffers.\n");
    return -3;
  }

  // set kernel args.
  unsigned int db_cnt = cell_count * cell_length / sizeof(DATA_TYPE);
  unsigned int db_cell_cnt = cell_length / sizeof(DATA_TYPE);
  err = 0;
  err  = clSetKernelArg(pipeline->kernel, 0, sizeof(cl_mem), &gpu_db);
  err |= clSetKernelArg(pipeline->kernel, 1, sizeof(cl_mem), &pipeline->gpu_input);
  err |= clSetKernelArg(pipeline->kernel, 2, cell_length, NULL);
  err |= clSetKernelArg(pipeline->kernel, 3, sizeof(unsigned int), &db_cnt);
  err |= clSetKernelArg(pipeline->kernel, 4, sizeof(unsigned int), &db_cell_cnt);
  err |= clSetKernelArg(pipeline->kernel, 5, sizeof(cl_mem), &pipeline->gpu_output);
  if (err != CL_SUCCESS)
  {
      printf("Error: Failed to set kernel arguments! %d\n", err);
      return -4;
  }

  return 0;
}

int pipeline_free(pipeline_t* pipeline) {
  clReleaseMemObject(pipeline->input_pin);
  clReleaseMemObject(pipeline->output_pin);
  clReleaseMemObject(pipeline->gpu_input);
  clReleaseMemObject(pipeline->gpu_output);
  clReleaseKernel(pipeline->kernel);

  return 0;
}

int pipeline_enqueue(pipeline_t* pipeline, int sock) {
  int err = 0;

  if (pipeline->input_loaded == TRUE) {
    return -2;
  }

  // Read input into host-mapped memory.
  int readamount = 0;
  while (readamount < cell_count * batch_size / 8) {
    err = read(sock, pipeline->input + readamount, cell_count * batch_size / 8 - readamount);
    if (err < 0) {
      return err;
    }
    readamount += err;
  }

  // enqueue transfer to gpu.
  err = clEnqueueWriteBuffer(commands,
                             pipeline->gpu_input,
                             CL_FALSE,
                             0,
                             cell_count * batch_size / 8,
                             pipeline->input,
                             0,
                             NULL,
                             NULL);
  if (err != CL_SUCCESS) {
    printf("Failed to transfer input to GPU!\n");
    return -1;
  }
  pipeline->input_loaded = TRUE;

  return 0;
}

int pipeline_dequeue(pipeline_t* pipeline, int sock) {
  int err = 0;
  int readamount = 0;
  if (pipeline->output_dirty == TRUE) {
    err = clEnqueueReadBuffer(commands,
                              pipeline->gpu_output,
                              CL_TRUE,
                              0,
                              cell_length * batch_size,
                              pipeline->output,
                              0,
                              NULL,
                              NULL);
    if (err != CL_SUCCESS) {
      printf("Failed to read response from GPU!\n");
      return -1;
    }

    // mark that we need to write to socket, but enqueue the next gpu
    // next, and then return to clearing the host buffer.
    readamount = 1;
    pipeline->output_dirty = FALSE;
  }

  if (pipeline->input_loaded == TRUE) {
    size_t totaljobs = workgroup_size * batch_size;
    err = clEnqueueNDRangeKernel(commands,
                                 pipeline->kernel,
                                 CL_FALSE,
                                 NULL,
                                 &totaljobs,
                                 &workgroup_size,
                                 0,
                                 NULL,
                                 NULL);
    if (err != CL_SUCCESS) {
      printf("Error: failed to execute pipeline computation!\n");
      return -1;
    }
    pipeline->output_dirty = TRUE;
    pipeline->input_loaded = FALSE;
  }


  // this is the return to clearing the host buffer, triggereed by output_dirty
  // being set in the first part of this function.
  if (readamount > 0) {
    readamount = 0;
    while (readamount < cell_length * batch_size) {
      err = write(sock, pipeline->output + readamount, cell_length * batch_size - readamount);
      if (err < 0) {
        printf("Failed to write response to client!\n");
        return err;
      }
      readamount += err;
    }
  }

  return 0;
}


int main(int argc, char** argv)
{
  int err;
  int c;
  int listonly = 0;
  socketpath = SOCKET_NAME;
  int deviceid = 0;
  while ((c = getopt(argc, argv, "s:d:l")) != -1) {
    switch (c) {
      case 'l':
        listonly = 1;
        break;
      case 'd':
        deviceid = atoi(optarg);
        break;
      case 's':
        socketpath = optarg;
        break;
      default:
        fprintf(stderr, "Usage: %s [-l] [-d <device id>] [-s <socket>].\n",
            argv[0]);
        return 1;
    }
  }

  if (listonly) {
    listDevices();
    return 1;
  }

  signal(SIGINT, intHandler);

  int socket_fd = socket(AF_UNIX, SOCK_STREAM, 0);
  int client_sock;
  if (socket_fd == -1) {
    printf("Error: Failed to create sockets!\n");
    return EXIT_FAILURE;
  }
  pipelines = malloc(sizeof(pipeline_t) * 2);

  // Bind the socket for communication.
  unlink((char*)socketpath);
  struct sockaddr_un socket_name;
  memset(&socket_name, 0, sizeof(struct sockaddr_un));
  socket_name.sun_family = AF_UNIX;
  strncpy(socket_name.sun_path, (char*)socketpath, sizeof(socket_name.sun_path) - 1);
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
  int next_pipeline = 0;
  int ret;
  int dbhndl;
  int newdbhndl;
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
        if (pipeline_enqueue(&pipelines[next_pipeline], client_sock) != 0) {
          printf("Failed to enqueue to pipeline.\n");
          break;
        }
        if (pipeline_dequeue(&pipelines[next_pipeline], client_sock) != 0) {
          printf("Failed to dequeue from pipeline.\n");
          break;
        }
        next_pipeline ^= 1;
      } else if (next_command == '2') { //configure
        ret = read(client_sock, configuration_params, 3*sizeof(int));
        if (ret == -1) {
          printf("Failed to read configuration.\n");
          break;
        }
        configure(deviceid,
                  configuration_params[0],
                  configuration_params[1],
                  configuration_params[2],
                  client_sock);
      } else if (next_command == '3') { // write
        ret = read(client_sock, &newdbhndl, sizeof(newdbhndl));
        if (ret == -1) {
          printf("Failed to learn shm id.\n");
          break;
        }
        if (newdbhndl != dbhndl && dbhndl != 0) {
          shmdt(database);
        }
        if (newdbhndl != dbhndl) {
          dbhndl = newdbhndl;
          database = shmat(dbhndl, NULL, SHM_RDONLY);
          if (database == (void*)-1) {
            printf("Failed to open shm ptr: %d.\n", errno);
            break;
          }
        }
        do_write(database);
        write(client_sock, "ok", 2);
      } else {
        printf("Unexpected command: %c\n", next_command);
        break;
      }
    }
  }
}

void listDevices() {
  int i, num;
  char buf[256];
  cl_device_id device_ids[10];
  cl_platform_id cl_platform;

  // Connect to a compute device
  int err = clGetPlatformIDs(1, &cl_platform, NULL);
  err = clGetDeviceIDs(cl_platform, CL_DEVICE_TYPE_ALL, 10, (cl_device_id*)&device_ids, &num);
  if (err != CL_SUCCESS)
  {
      printf("Error: Failed get device IDs!\n");
      return;
  }
  for (i = 0; i < num && device_ids[i] != 0; i++) {
    err = clGetDeviceInfo(device_ids[i], CL_DEVICE_NAME, 256, &buf, NULL);
    if (err != CL_SUCCESS) {
      printf("%d: <Failed to read name.>\n", i);
      continue;
    }
    printf("%d: %s\n", i, buf);
  }
}

int configure(int devid, int n_cell_length, int n_cell_count, int n_batch_size, int client_sock) {
  int err;

  cell_length = n_cell_length;
  cell_count = n_cell_count;
  batch_size = n_batch_size;

  if (pipelines[0].input_loaded == TRUE|| pipelines[0].output_dirty == TRUE) {
    // Stall until in-process stuff is done.
    clFinish(commands);
	  // respond to pending reads.
    pipeline_dequeue(&pipelines[0], client_sock);
    pipeline_dequeue(&pipelines[0], client_sock);
    pipeline_dequeue(&pipelines[1], client_sock);
    pipeline_dequeue(&pipelines[1], client_sock);
    pipeline_free(&pipelines[0]);
    pipeline_free(&pipelines[1]);
  }

  if (context != NULL) {
    clReleaseProgram(program);
    clReleaseCommandQueue(commands);
    clReleaseContext(context);
  }

  // Set the socket buffer to as much of a capacity as we can.
  int maxbuf = cell_count * batch_size / 8;
  if (cell_length * batch_size > maxbuf) {
    maxbuf = cell_length * batch_size;
  }
  maxbuf *= 2;
  int buflen = sizeof(maxbuf);
  setsockopt(client_sock, SOL_SOCKET, SO_RCVBUF, &maxbuf, buflen);
  setsockopt(client_sock, SOL_SOCKET, SO_SNDBUF, &maxbuf, buflen);

  cl_device_id device_id[10];             // compute device id
  cl_platform_id cl_platform;

  // Connect to a compute device
  err = clGetPlatformIDs(1, &cl_platform, NULL);
  err = clGetDeviceIDs(cl_platform, CL_DEVICE_TYPE_ALL, 10, (cl_device_id*)&device_id, NULL);
  if (err != CL_SUCCESS)
  {
      printf("Error: Failed to create a device group!\n");
      return EXIT_FAILURE;
  }

  // Create a compute context
  context = clCreateContext(0, 1, &device_id[devid], NULL, NULL, &err);
  if (!context)
  {
      printf("Error: Failed to create a compute context!\n");
      return EXIT_FAILURE;
  }

  // Create a command queue
  commands = clCreateCommandQueue(context, device_id[devid], 0, &err);
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
      clGetProgramBuildInfo(program, device_id[devid], CL_PROGRAM_BUILD_LOG, sizeof(buffer), buffer, &len);
      printf("%s\n", buffer);
      exit(1);
  }

  // set up pipelines.
  pipeline_init(&pipelines[0]);
  pipeline_init(&pipelines[1]);

  // learn about the device memory size:
  //err = clGetDeviceInfo(device_id, CL_DEVICE_MAX_MEM_ALLOC_SIZE, sizeof(cl_ulong), &buf_max_size, NULL);
  //err = clGetDeviceInfo(device_id, CL_DEVICE_GLOBAL_MEM_SIZE, sizeof(cl_ulong), &dev_mem_size, NULL);
  //printf("Global size is %lu. Maximum buffer on device can be %lu.\n", dev_mem_size, buf_max_size);

  // Get the maximum work group size for executing the kernel on the device
  //
  err = clGetKernelWorkGroupInfo(pipelines[0].kernel, device_id[devid], CL_KERNEL_WORK_GROUP_SIZE, sizeof(workgroup_size), &workgroup_size, NULL);
  if (err != CL_SUCCESS)
  {
      printf("Error: Failed to retrieve kernel work group info! %d\n", err);
      exit(1);
  }

  if (gpu_db != NULL) {
    clReleaseMemObject(gpu_db);
  }
  gpu_db = clCreateBuffer(context,
                          CL_MEM_READ_ONLY,
                          cell_length * cell_count,
                          NULL,
                          NULL);

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
