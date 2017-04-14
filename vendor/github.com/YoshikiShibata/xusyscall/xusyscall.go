// Copyright (c) 2013 Yoshiki Shibata. All rights reserved.

// Package xusyscall contains system calls which are not
// provided the original Go syscall package.
package xusyscall

// Note that some system call such as shmget is not a system call on some
// systems such as 32bit Linux. So we need to implement this package in C
// language

// #include <sys/ipc.h>
// #include <sys/shm.h>
// #include <errno.h>
import "C"

import "unsafe"
import "strconv"

// Flags for fuctions. About details see the output of `man shmget` on Unix.
const (
	// Private key
	IPC_PRIVATE = 0

	// Create key if key does not exist
	IPC_CREAT = 01000
	// Fail if key exists
	IPC_EXCL = 02000

	// Read-Only access
	shm_RDONLY = 010000

	ipc_RMID = 0
	ipc_SET  = 1
	ipc_STAT = 2
	ipc_INFO = 3
)

// Gets a shared memory specified by the key.
// key, size, and shmflg must be equal to or greater than 0.
func Shmget(key int, size int, shmflg int) (shmid int, err error) {
	if key < 0 {
		panic("key is negative value: " + strconv.Itoa(key))
	}
	if size < 0 {
		panic("size is negative value: " + strconv.Itoa(size))
	}
	if shmflg < 0 {
		panic("shmflg is negative value: " + strconv.Itoa(shmflg))
	}

	result, errno := C.shmget(C.key_t(key), C.size_t(size), C.int(shmflg))

	if result == -1 {
		return -1, errno
	}
	return int(result), nil
}

// Attaches the specified shared memory.
// readOnly is used to attach the memory in read-only mode.
func Shmat(shmid int, readOnly bool) (data []byte, err error) {
	var shmflg = 0;

	if readOnly {
		shmflg = shm_RDONLY
	}
	
	addr, errno := shmat(shmid, 0, shmflg)

	if errno != nil {
		return nil, errno
	}

	length, errno2 := shmseqsz(shmid)
	if errno2 != nil {
		return nil, errno2
	}

	// Slice memory layout: see the implementation of Mmap
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{addr, length, length}

	data = *(*[]byte)(unsafe.Pointer(&sl))
	return data, nil
}

func shmat(shmid int, shmaddr uintptr, shmflg int) (addr uintptr, err error) {
	result, errno := C.shmat(C.int(shmid), unsafe.Pointer(shmaddr), C.int(shmflg))

	if int(uintptr(result)) == -1 {
		return 0, errno
	}

	return uintptr(result), nil
}

func shmseqsz(shmid int) (segsz int, err error) {
	var shmid_ds C.struct_shmid_ds

	errno := shmctl(shmid, ipc_STAT, &shmid_ds)
	if errno != nil {
		return 0, errno
	}
	return int(shmid_ds.shm_segsz), nil
}

// Detaches the shared memory
func Shmdt(data []byte) error {
	result, errno := C.shmdt(unsafe.Pointer(&data[0]))

	if result == -1 {
		return errno
	}
	return nil
}

// Remove the shared memory specified by shmid
func Shmrm(shmid int) error {
	var shmid_ds C.struct_shmid_ds

	errno := shmctl(shmid, ipc_RMID, &shmid_ds)

	if errno != nil {
		return errno
	}
	return nil
}

// shmctl syscall
func shmctl(shmid int, cmd int, shmid_ds *C.struct_shmid_ds) error {
	result, errno := C.shmctl(C.int(shmid), C.int(cmd),
		(*C.struct_shmid_ds)(unsafe.Pointer(shmid_ds)))

	if result == -1 {
		return errno
	}
	return nil
}
