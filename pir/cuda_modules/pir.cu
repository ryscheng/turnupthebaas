/***************************************************
* Module for PIR
*
* To be compiled with nvcc -ptx pir.cu
* Debug: nvcc -arch=sm_20 -ptx pir.cu
**************************************************/

//#include <stdio.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef long int int32_cu
typedef unsigned long int uint32_cu
typedef long long int int64_cu
typedef unsigned long long int uint64_cu
#define DATA_TYPE uint64_cu

// CUDA Kernel
__global__
void pir(DATA_TYPE* db,
        char* reqs,
        DATA_TYPE* output,
        DATA_TYPE* scratch,
        uint32_cu batchSize,
	      uint32_cu reqLength,
        uint32_cu numBuckets,
        uint32_cu bucketSize,
        uint32_cu globalSize,
        uint32_cu scratchSize) {
  int localIndex = threadIdx.x;
  int groupIndex = blockIdx.x;
  int globalIndex = threadIdx.x + (blockIdx.x * blockDim.x);

  if (globalIndex >= globalSize) {
    return;
  }
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

#ifdef __cplusplus
}
#endif
