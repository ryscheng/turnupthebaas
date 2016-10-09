__kernel
void pir(__global unsigned long* buffer,
         __global char* mask,
         __local unsigned long* scratch,
         __const int length,                 // total number of longs in buffer.
         __const int cell_length,            // number of longs in 1 'cell'
         __global unsigned long* output) {

  int workgroup_size = get_local_size(0);
  int workgroup_index = get_local_id(0);
  int workgroup_num = get_group_id(0);
  int mask_offset = workgroup_num * (length / cell_length);

  // zero scratch
  for (int offset = workgroup_index; offset < cell_length; offset += workgroup_size) {
      scratch[offset] = 0;
  }
  barrier(CLK_LOCAL_MEM_FENCE);

  // Accumulate in parallel.
  for (int offset = workgroup_index; offset < length; offset += workgroup_size) {
    if (mask[mask_offset + offset / cell_length]) {
      scratch[offset % cell_length] ^= buffer[offset];
    }
  }

  // send to output.
  barrier(CLK_LOCAL_MEM_FENCE);
  if (workgroup_index == 0) {
    for (int offset = 0; offset < cell_length; offset += 1) {
      output[workgroup_num * cell_length + offset] = scratch[offset];
    }
  }
}
