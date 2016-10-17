package pircl

import "bytes"
import "C"
import "encoding/binary"
import "errors"
import "fmt"
import "math/rand"
import "net"
import "syscall"
import "unsafe"

type PirDB struct {
  DB      []byte
  dbptr   *byte
  shmid   uintptr
}

type PirServer struct {
	sock       net.Conn
	CellLength int
	CellCount  int
	BatchSize  int
  DB         *PirDB
}

const pirCommands = `123`
const defaultSocket = "pir.socket"

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

func destroyShm(mem *byte) error {
	addr := unsafe.Pointer(mem)
	ret, _, errno := syscall.RawSyscall(syscall.SYS_SHMDT, uintptr(addr), 0, 0)
	if ret != 0 || errno != 0 {
    return errors.New("Failed to release SHM: " + string(errno))
	}
	return nil
}

func Connect(socket string) (PirServer, error) {
  if socket == nil {
    socket = defaultSocket
  }

  sock, err := net.Dial("unix", socket)
  if err != nil {
    return nil, err
  }

  server := make(PirServer)
  server.sock = sock
  return server, nil
}

func (s *PirServer) Configure(celllength int, cellcount int, batchsize int) error {
  s.BatchSize = batchsize
  s.CellCount = cellcount
  s.CellLength = celllength

  buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(cellength))
	binary.Write(buf, binary.LittleEndian, int32(cellcount))
	binary.Write(buf, binary.LittleEndian, int32(batchsize))
	_, err := s.sock.Write([]byte{pirCommands[1]})
  if err != nil {
    return err
  }
	_, err = s.sock.Write(buf.Bytes())
	if err != nil {
    return err
	}
  return nil
}

func (s *PirServer) GetDB() (*PirDB, error) {
  dbptr, database := createShm(s.CellLength * s.CellCount)
	if database == nil {
    return nil, errors.New("Could not create shared memory for database")
	}
  dbslice := C.GoBytes(unsafe.Pointer(database), C.int(s.CellLength * s.CellCount))
  return dbslice, nil
}

func (s *PirServer) SetDB(db *PirDB) error {
  if _, err = s.sock.Write([]byte{pirCommands[2]}); err != nil {
    return err
  }

  buf = new(bytes.Buffer)
  binary.Write(buf, binary.LittleEndian, int32(db.shmid))
  _, err = sock.Write(buf.Bytes())
  if err != nil {
    return err
  }
  s.DB = db
  return nil
}

func (db *PirDB) Free() error {
  return destroyShm(db.dbptr)
}

func (s *PirServer) Read(masks []byte) ([]byte, error) {
  if s.DB == nil || s.CellCount == 0 {
    return nil, errors.New("DB not configured.")
  }

  if len(masks) != (s.CellLength * s.BatchSize) / 8 {
    return nil, errors.New("Wrong Mask Length.")
  }

  if _, err := s.sock.Write([]byte{pirCommands[0]}); err != nil {
    return nil, err
  }

  if _, err := s.sock.Write(masks); err != nil {
    return nil, err
  }

  responses := make([]byte, s.CellLength * s.BatchSize)
  readNum := 0
  for readNum < len(responses) {
    count, err := s.sock.Read(responses[readNum:])
    readNum += count
    if count <= 0 || err != nil {
      return nil, err
    }
  }
  return responses, nil
}
