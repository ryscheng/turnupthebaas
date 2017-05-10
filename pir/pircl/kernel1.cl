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

}
