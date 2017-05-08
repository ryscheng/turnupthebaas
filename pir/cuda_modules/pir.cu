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
        int batchSize,
        int reqLength,
        int numBuckets,
        int bucketSize,
        int globalSize){
  //int localIndex = threadIdx.x;
  //int groupIndex = blockIdx.x;
  int globalIndex = threadIdx.x + (blockIdx.x * blockDim.x);

  if (globalIndex >= globalSize) {
    return;
  }
  __syncthreads();

  // Iterate over requests in a batch, atomic_xor my data into output
  int bucketId = globalIndex / bucketSize;
  int depthOffset = globalIndex % bucketSize;
  DATA_TYPE data = db[globalIndex];
  DATA_TYPE* addr;
  uint8_cu reqBit;
  for (int i = 0; i < batchSize; i++) {
    reqBit = reqs[(i*reqLength) + (bucketId/8)] & (1 << (bucketId%8));
    if (reqBit > 0) {
      addr = &output[(i*bucketSize)+depthOffset];
      atomicXor(addr, data);
    }
  }
}

#ifdef __cplusplus
}
#endif
