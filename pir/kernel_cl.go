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
	// KernelDataSize must correspond to DATA_TYPE in the kernel
	KernelDataSize = 8
)

const kernelCLPrefix = `
typedef char int8_cl;
typedef unsigned char uint8_cl;
typedef int int32_cl;
typedef unsigned int uint32_cl;
typedef long int64_cl;
typedef unsigned long uint64_cl;
#pragma OPENCL EXTENSION cl_khr_int64_extended_atomics : enable
#define DATA_TYPE uint64_cl
__kernel
void pir(__global DATA_TYPE* db,
	__global uint8_cl* reqs,
        __global DATA_TYPE* output,
        __local DATA_TYPE* scratch,
        __const uint32_cl batchSize,
	__const uint32_cl reqLength,
	__const uint32_cl numBuckets,
	__const uint32_cl bucketSize,
	__const uint32_cl globalSize,
	__const uint32_cl scratchSize) {
  //uint32_cl globalSize = get_global_size(0);
  //uint32_cl localSize = get_local_size(0);
  //uint32_cl localIndex = get_local_id(0);
  //uint32_cl groupIndex = get_group_id(0);
  //uint32_cl globalIndex = get_global_id(0);
`

const kernelCLPostfix = `
}
` + "\x00"

// KernelCL0 : 1 workgroup == 1 PIR request
// Workgroup items split up the scan over the database
// Cache the working result
const KernelCL0 = kernelCLPrefix + `
  uint32_cl workgroup_size = get_local_size(0);
  uint32_cl workgroup_index = get_local_id(0);
  uint32_cl workgroup_num = get_group_id(0);	  // request index

  // zero scratch
  for (uint32_cl offset = workgroup_index; offset < bucketSize; offset += workgroup_size) {
      scratch[offset] = 0;
  }
  barrier(CLK_LOCAL_MEM_FENCE);

  // Accumulate in parallel.
  uint32_cl dbSize = numBuckets * bucketSize;
  uint32_cl reqIndex = workgroup_num * reqLength;
  uint32_cl bucketId;
  uint32_cl depthOffset;
  uint8_cl reqBit;
  for (uint32_cl offset = workgroup_index; offset < dbSize; offset += workgroup_size) {
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
  uint32_cl respIndex = workgroup_num * bucketSize;
  for (uint32_cl offset = workgroup_index; offset < bucketSize; offset += workgroup_size) {
    output[respIndex + offset] = scratch[offset];
  }
` + kernelCLPostfix

// KernelCL1 : index => output
// Cache the request
const KernelCL1 = kernelCLPrefix + `
  uint32_cl globalIndex = get_global_id(0);

  if (globalIndex >= globalSize) {
    return;
  }

  //barrier(CLK_LOCAL_MEM_FENCE);
  
  // Iterate over all buckets, xor data into my result
  DATA_TYPE result = 0;
  uint32_cl reqIndex = (globalIndex / bucketSize) * reqLength;
  uint32_cl offset = globalIndex % bucketSize;
  uint8_cl reqBit;
  for (uint32_cl i = 0; i < numBuckets; i++) {
    reqBit = reqs[reqIndex + (i/8)] & (1 << (i%8));
    if (reqBit > 0) {
      result ^= db[i*bucketSize+offset];
    }
  }
  output[globalIndex] = result;
` + kernelCLPostfix

// KernelCL2 : index => output
// Cache a portion of the database
const KernelCL2 = kernelCLPrefix + `
  uint32_cl globalIndex = get_global_id(0);

  if (globalIndex >= globalSize) {
    return;
  }

  //barrier(CLK_LOCAL_MEM_FENCE);
  
  // Iterate over all buckets, xor data into my result
  DATA_TYPE result = 0;
  uint32_cl reqIndex = (globalIndex / bucketSize) * reqLength;
  uint32_cl offset = globalIndex % bucketSize;
  uint8_cl reqBit;
  for (uint32_cl i = 0; i < numBuckets; i++) {
    reqBit = reqs[reqIndex + (i/8)] & (1 << (i%8));
    if (reqBit > 0) {
      result ^= db[i*bucketSize+offset];
    }
  }
  output[globalIndex] = result;
` + kernelCLPostfix

// KernelCL3 : index => db
// Cache portion of the database
const KernelCL3 = kernelCLPrefix + `
  uint32_cl globalIndex = get_global_id(0);

  if (globalIndex >= globalSize) {
    return;
  }

  uint32_cl outputSize = batchSize * bucketSize;

  // Zero output first
  if (globalSize >= outputSize && globalIndex < outputSize) {
    output[globalIndex] = 0;
  } else if (globalSize < outputSize) {
    uint32_cl multiplier = outputSize / globalSize + 1;
    uint32_cl start = globalIndex * multiplier;
    uint32_cl end = start + multiplier;
    for (uint32_cl i = start; i < end && i < outputSize; i++) {
      output[i] = 0;
    }
  }
  barrier(CLK_GLOBAL_MEM_FENCE);

  // Iterate over requests in a batch, atomic_xor my data into output
  DATA_TYPE data = db[globalIndex];
  uint32_cl bucketId = globalIndex / bucketSize;
  uint32_cl depthOffset = globalIndex % bucketSize;
  uint8_cl reqBit;
  for (uint32_cl i = 0; i < batchSize; i++) {
    reqBit = reqs[(i*reqLength) + (bucketId/8)] & (1 << (bucketId%8));
    if (reqBit > 0) {
      atom_xor(&output[(i*bucketSize)+depthOffset], data);
    }
  }
` + kernelCLPostfix
