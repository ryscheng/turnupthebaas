package main

import "bytes"
import "C"
import "encoding/binary"
import "fmt"
import "math/rand"
import "net"
import "syscall"
import "unsafe"


//from http://golangtc.com/t/531072f4320b5261970000ba
func createShm(size int) (shmid uintptr, mem *byte) {
  flag := 0600
  shmid, _, errno := syscall.RawSyscall(syscall.SYS_SHMGET, 0, uintptr(size), uintptr(flag))

  addr, _, errno := syscall.RawSyscall(syscall.SYS_SHMAT, shmid, 0, 0)
  mem = (*byte)(unsafe.Pointer(addr))
  if errno != 0 {
    fmt.Printf("Failed to create SHM: %d", errno)
  }
  return
}

func destroyShm(mem *byte) {
  addr := unsafe.Pointer(mem)
  ret, _, errno := syscall.RawSyscall(syscall.SYS_SHMDT, uintptr(addr), 0, 0)
  if ret != 0 || errno != 0 {
    fmt.Printf("Failed to release SHM: %d", errno)
  }
  return
}

func main() {
  sock, err := net.Dial("unix", "pir.socket")
  if err != nil {
    panic("Failed to connect to socket.")
  }
  fmt.Println("Connected to PIR Daemon.")

  cell_length := int32(1024)
  cell_count := int32(1024)
  batch_size := int32(8)
  commands := `123`

  buf := new(bytes.Buffer)
  binary.Write(buf, binary.LittleEndian, cell_length)
  binary.Write(buf, binary.LittleEndian, cell_count)
  binary.Write(buf, binary.LittleEndian, batch_size)
  _, err = sock.Write([]byte{commands[1]})
  _, err = sock.Write(buf.Bytes())
  if err != nil {
    panic("Failed to write configuration.")
  }
  fmt.Printf("Configured PIR to %d cells of %d bytes and read batches of %d\n", cell_length, cell_count, batch_size)

  db_fp, database := createShm(int(cell_length * cell_count))
  if database == nil {
    fmt.Printf("shm open errored %d", database)
    panic("Could not create shared memory for database")
  }
  defer destroyShm(database)

  dbslice := C.GoBytes(unsafe.Pointer(database), C.int(cell_length * cell_count))
  cell := make([]byte, cell_length)
  rand.Read(cell)
  for i:=0; i < int(cell_count); i++ {
    copy(dbslice[int(cell_length)*i:], cell)
  }

  _, err = sock.Write([]byte{commands[2]})
  buf = new(bytes.Buffer)
  binary.Write(buf, binary.LittleEndian, int32(db_fp))
  _, err = sock.Write(buf.Bytes())
  if err != nil {
    panic("Failed to set database.")
  }
  fmt.Println("PIR Database set.\n")


  masks := make([]byte, cell_count * batch_size / 8)
  rand.Read(masks)

  _, err = sock.Write([]byte{commands[0]})
  _, err = sock.Write(masks)
  if err != nil {
    panic("Failed to ask for read.")
  }
  fmt.Println("Requested Read.\n")

  responses := make([]byte, cell_length * batch_size)
  readNum := 0
  for ; readNum < len(responses); {
    count, _ := sock.Read(responses[readNum:])
    readNum += count
  }
  fmt.Println("Received Response.\n")
}
