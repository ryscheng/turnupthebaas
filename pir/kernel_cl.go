package pir

// https://www.khronos.org/files/opencl-quick-reference-card.pdf

// db: shard ([bucket0, bucket1, ...]) where each bucket is bucketSize bytes
// reqs: batch of request vectors ([req0, req1, ...]) where each req is reqLength bytes
// output: batch of responses ([resp0, resp1, ...]) where each resp is bucketSize bytes
// scratch: L2 scratchpad of GPUScratchSize bytes
// batchSize: number of requests per batch
// reqLength: length of a request in bytes (numBuckets/8)
// numBuckets: number of buckets in the shard
// bucketSize: length of a bucket in units of DATA_TYPE
// globalSize: number of threads globally (size of db if Kernel0, size of output if Kernel1)
// scratchSize: length of scratch in units of DATA_TYPE

const (
	// GPUScratchSize is the size of GPU scratch/L1 cache in bytes
	GPUScratchSize = 2048
	// KernelDataSize must correspond to DATA_TYPE in the kernel
	KernelDataSize = 8
)

// KernelCL0 : 1 workgroup == 1 PIR request
// Workgroup items split up the scan over the database
// Cache the working result
const KernelCL0 = `
#pragma OPENCL EXTENSION cl_khr_int64_extended_atomics : enable
#define DATA_TYPE unsigned long
__kernel
void pir(__global DATA_TYPE* db,
	__global char* reqs,
        __global DATA_TYPE* output,
        __local DATA_TYPE* scratch,
        __const unsigned int batchSize,
	__const unsigned int reqLength,
	__const unsigned int numBuckets,
	__const unsigned int bucketSize,
	__const unsigned int globalSize,
	__const unsigned int scratchSize) {

  int workgroup_size = get_local_size(0);
  int workgroup_index = get_local_id(0);
  int workgroup_num = get_group_id(0);	  // request index

  // zero scratch
  for (int offset = workgroup_index; offset < bucketSize; offset += workgroup_size) {
      scratch[offset] = 0;
  }
  barrier(CLK_LOCAL_MEM_FENCE);

  // Accumulate in parallel.
  int dbSize = numBuckets * bucketSize;
  int reqIndex = workgroup_num * reqLength;
  int bucketId;
  int depthOffset;
  unsigned char reqBit;
  for (int offset = workgroup_index; offset < dbSize; offset += workgroup_size) {
    bucketId = offset / bucketSize;
    depthOffset = offset % bucketSize;
    reqBit = reqs[reqIndex + (bucketId/8)] & (1 << (bucketId%8));
    //current_mask = (current_mask >> bitshift) * -1;
    //scratch[depthOffset] ^= current_mask & db[offset];
    if (reqBit != 0) {
      //scratch[depthOffset] ^= db[offset];
      atom_xor(&scratch[depthOffset], db[offset]);
    }
  }

  // send to output.
  barrier(CLK_LOCAL_MEM_FENCE);
  int respIndex = workgroup_num * bucketSize;
  for (int offset = workgroup_index; offset < bucketSize; offset += workgroup_size) {
    output[respIndex + offset] = scratch[offset];
  }
}
` + "\x00"

// KernelCL1 : index => output
// Cache the request
const KernelCL1 = `
#define DATA_TYPE unsigned long
__kernel
void pir(__global DATA_TYPE* db,
        __global char* reqs,
        __global DATA_TYPE* output,
        __local DATA_TYPE* scratch,
        __const unsigned int batchSize,
	__const unsigned int reqLength,
	__const unsigned int numBuckets,
	__const unsigned int bucketSize,
	__const unsigned int globalSize,
	__const unsigned int scratchSize) {
  //int globalSize = get_global_size(0);
  int localSize = get_local_size(0);
  int localIndex = get_local_id(0);
  int groupIndex = get_group_id(0);
  int globalIndex = get_global_id(0);

  if (globalIndex >= globalSize) {
    return;
  }

  //barrier(CLK_LOCAL_MEM_FENCE);
  
  // Iterate over all buckets, xor data into my result
  DATA_TYPE result = 0;
  int reqIndex = (globalIndex / bucketSize) * reqLength;
  int offset = globalIndex % bucketSize;
  unsigned char reqBit;
  for (int i = 0; i < numBuckets; i++) {
    reqBit = reqs[reqIndex + (i/8)] & (1 << (i%8));
    if (reqBit > 0) {
      result ^= db[i*bucketSize+offset];
    }
  }
  output[globalIndex] = result;

}
` + "\x00"

// KernelCL2 : index => output
// Cache a portion of the database
const KernelCL2 = `
#define DATA_TYPE unsigned long
__kernel
void pir(__global DATA_TYPE* db,
	__global char* reqs,
        __global DATA_TYPE* output,
        __local DATA_TYPE* scratch,
        __const unsigned int batchSize,
	__const unsigned int reqLength,
	__const unsigned int numBuckets,
	__const unsigned int bucketSize,
	__const unsigned int globalSize,
	__const unsigned int scratchSize) {
  //int globalSize = get_global_size(0);
  int localSize = get_local_size(0);
  int localIndex = get_local_id(0);
  int groupIndex = get_group_id(0);
  int globalIndex = get_global_id(0);

  if (globalIndex >= globalSize) {
    return;
  }

  //barrier(CLK_LOCAL_MEM_FENCE);
  
  // Iterate over all buckets, xor data into my result
  DATA_TYPE result = 0;
  int reqIndex = (globalIndex / bucketSize) * reqLength;
  int offset = globalIndex % bucketSize;
  unsigned char reqBit;
  for (int i = 0; i < numBuckets; i++) {
    reqBit = reqs[reqIndex + (i/8)] & (1 << (i%8));
    if (reqBit > 0) {
      result ^= db[i*bucketSize+offset];
    }
  }
  output[globalIndex] = result;

}
` + "\x00"

// KernelCL3 : index => db
// Cache portion of the database
const KernelCL3 = `
#pragma OPENCL EXTENSION cl_khr_int64_extended_atomics : enable
#define DATA_TYPE unsigned long
__kernel
void pir(__global DATA_TYPE* db,
	__global char* reqs,
        __global DATA_TYPE* output,
        __local DATA_TYPE* scratch,
        __const unsigned int batchSize,
	__const unsigned int reqLength,
	__const unsigned int numBuckets,
	__const unsigned int bucketSize,
	__const unsigned int globalSize,
	__const unsigned int scratchSize) {
  //int globalSize = get_global_size(0);
  int localSize = get_local_size(0);
  int localIndex = get_local_id(0);
  int groupIndex = get_group_id(0);
  int globalIndex = get_global_id(0);

  if (globalIndex >= globalSize) {
    return;
  }

  int outputSize = batchSize * bucketSize;

  // Zero output first
  if (globalSize >= outputSize && globalIndex < outputSize) {
    output[globalIndex] = 0;
  } else if (globalSize < outputSize) {
    int multiplier = outputSize / globalSize + 1;
    int start = globalIndex * multiplier;
    int end = start + multiplier;
    for (int i = start; i < end && i < outputSize; i++) {
      output[i] = 0;
    }
  }
  barrier(CLK_GLOBAL_MEM_FENCE);

  // Iterate over requests in a batch, atomic_xor my data into output
  DATA_TYPE data = db[globalIndex];
  int bucketId = globalIndex / bucketSize;
  int depthOffset = globalIndex % bucketSize;
  unsigned char reqBit;
  for (int i = 0; i < batchSize; i++) {
    reqBit = reqs[(i*reqLength) + (bucketId/8)] & (1 << (bucketId%8));
    if (reqBit > 0) {
      atom_xor(&output[(i*bucketSize)+depthOffset], data);
    }
  }

}
` + "\x00"
