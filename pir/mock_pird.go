package pir

import "encoding/binary"
import "fmt"
import "net"
import "unsafe"
import "github.com/YoshikiShibata/xusyscall"

// This package provides the functionality of the PIR Daemon - it opens a
// socket, and uses the same protocol to communicate.
// Notably: It does not make use of GPU accelleration for PIR computation, so
// will be much slower. It does, however, make the prospect of testing the
// code base much simpler.

func CreateMockServer(status chan int, socket string) error {
	if len(socket) == 0 {
		socket = defaultSocket
	}

	sock, err := net.Listen("unix", socket)
	if err != nil {
		status <- -1
		return err
	}

	activeConn := make([]net.Conn, 1)

	go quitWatchdog(status, &sock, activeConn)
	status <- 1

	for {
		conn, err := sock.Accept()
		if err != nil {
			return err
		}
		activeConn[0] = conn
		handle(conn)
	}
}

func quitWatchdog(channel chan int, listener *net.Listener, active []net.Conn) {
	<- channel
	(*listener).Close()
	if active[0] != nil {
		active[0].Close()
	}
}

func handle(conn net.Conn) {
	var database []byte

	CellLength := int(1024)
	CellCount := int(1024)
	BatchSize := int(8)

	// handle connection.
	for {
		// read first byte
		cmd := make([]byte, 1)
		if len, err := conn.Read(cmd); len < 1 || err != nil {
			break
		}
		if cmd[0] == "1"[0] {
			// read.
			// read mask
			masks := make([]byte, CellCount*BatchSize/8)
			if len, err := conn.Read(masks); len < CellLength*BatchSize/8 || err != nil {
				break
			}

			// calculate pir
			response := make([]byte, CellLength*BatchSize)
			for batch := 0; batch < BatchSize; batch += 1 {
				for cell := 0; cell < CellCount; cell += 1 {
					if (masks[(batch*CellCount+cell)/8] & (1 << uint(cell%8))) != 0 {
						for off := 0; off < CellLength; off += 1 {
							response[batch*CellLength+off] ^= database[CellLength*cell+off]
						}
					}
				}
			}
			if database[0] == 1 {
				fmt.Printf("Databse looks set correctly.\n")
			}

			// write response
			writeNum := 0
			for writeNum < len(response) {
				count, err := conn.Write(response[writeNum:])
				writeNum += count
				if count <= 0 || err != nil {
					break
				}
			}
		} else if cmd[0] == "2"[0] {
			// configure.
			cfgs := make([]byte, 12)
			conn.Read(cfgs)
			CellLength = int(binary.LittleEndian.Uint32(cfgs[0:4]))
			CellCount = int(binary.LittleEndian.Uint32(cfgs[4:8]))
			BatchSize = int(binary.LittleEndian.Uint32(cfgs[8:12]))
		} else if cmd[0] == "3"[0] {
			// write.

			// read shared memory ID
			var err error
			shmidarr := make([]byte, unsafe.Sizeof(int(0)))
			conn.Read(shmidarr)
			shmid := binary.LittleEndian.Uint32(shmidarr)
			// attach shared memory
			database, err = xusyscall.Shmat(int(shmid), true)
			if err != nil {
				fmt.Printf("Couldn't read database from ptr. %s", err)
				conn.Close()
				break
			}
		}
	}
}
