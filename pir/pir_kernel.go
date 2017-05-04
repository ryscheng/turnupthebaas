package pir

const KernelSource = `
__kernel
void pir(__global unsigned long* db,
         __global char* reqs,
         __global unsigned long* output) {
  int workgroup_size = get_local_size(0);
  int workgroup_index = get_local_id(0);
  int workgroup_num = get_group_id(0);
}
` + "\x00"

const KernelSource1 = `
__kernel
void pir(__global unsigned long* db,
         __global char* reqs,
         __global unsigned long* output,
         __local unsigned long* scratch,
         __const int length,                 // total number of longs in db.
         __const int cell_length) { // number of longs in 1 'cell'

  int workgroup_size = get_local_size(0);
  int workgroup_index = get_local_id(0);
  int workgroup_num = get_group_id(0);
  int mask_offset = workgroup_num * (length / cell_length) / 8;
  long current_mask;
  short bitshift;

  // zero scratch
  for (int offset = workgroup_index; offset < cell_length; offset += workgroup_size) {
      scratch[offset] = 0;
  }
  barrier(CLK_LOCAL_MEM_FENCE);

  // Accumulate in parallel.
  for (int offset = workgroup_index; offset < length; offset += workgroup_size) {
    bitshift = offset / cell_length % 8;
    current_mask = reqs[mask_offset + offset / cell_length / 8] & (1 << bitshift);
    current_mask = (current_mask >> bitshift) * -1;
    scratch[offset % cell_length] ^= current_mask & db[offset];
  }

  // send to output.
  barrier(CLK_LOCAL_MEM_FENCE);
  for (int offset = workgroup_index; offset < cell_length; offset += workgroup_size) {
    output[workgroup_num * cell_length + offset] = scratch[offset];
  }
}
` + "\x00"
