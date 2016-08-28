__kernel
void pir(__global unsigned long* buffer,
         __global char* mask,
         __local unsigned long* scratch,
         __const int length,
         __const int scratch_length,
         __global unsigned long* output) {

  int global_index = get_global_id(0);
  // Loop sequentially over chunks of input vector
  while (global_index < length) {
    if (mask[global_index]) {
      scratch[global_index % scratch_length] ^= buffer[global_index];
    }
    global_index += get_global_size(0);
  }

  // Perform parallel reduction
  int local_index = get_local_id(0);
  barrier(CLK_LOCAL_MEM_FENCE);
  if (local_index == 0) {
    for (int offset = 0; offset < scratch_length; offset ++) {
      output[offset] ^= scratch[offset];
    }
  }
}
