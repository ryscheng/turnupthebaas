/***************************************************
* Module for PIR
*
* To be compiled with nvcc -ptx pir.cu
* Debug: nvcc -arch=sm_20 -ptx pir.cu
* Note: CUDA may not support all versions of gcc;
* See
* https://groups.google.com/forum/#!topic/torch7/WaNmWZqMnzw
**************************************************/

//#include <stdio.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef char int8_cu;
typedef unsigned char uint8_cu;
typedef long int int32_cu;
typedef unsigned long int uint32_cu;
typedef long long int int64_cu;
typedef unsigned long long int uint64_cu;
#define DATA_TYPE uint64_cu

// CUDA Kernel
__global__
void pir(DATA_TYPE* db,
        uint8_cu* reqs,
        DATA_TYPE* output,
        //DATA_TYPE* scratch,
        uint32_cu batchSize,
        uint32_cu reqLength,
        uint32_cu numBuckets,
        uint32_cu bucketSize,
        uint32_cu globalSize){
        //uint32_cu scratchSize) {
  //int localIndex = threadIdx.x;
  //int groupIndex = blockIdx.x;
  uint32_cu globalIndex = threadIdx.x + (blockIdx.x * blockDim.x);

  if (globalIndex >= globalSize) {
    return;
  }
  // Iterate over all buckets, xor data into my result
  DATA_TYPE result = 0;
  uint32_cu reqIndex = (globalIndex / bucketSize) * reqLength;
  uint32_cu offset = globalIndex % bucketSize;
  uint8_cu reqBit;
  for (uint32_cu i = 0; i < numBuckets; i++) {
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
