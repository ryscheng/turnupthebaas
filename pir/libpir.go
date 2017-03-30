package pir

import "bytes"
import "encoding/binary"
import "errors"
import "fmt"
import "net"
import "github.com/YoshikiShibata/xusyscall"

// DB is a memory area for PIR computations shared with a PIR daemon.
type DB struct {
	DB    []byte
	shmid int
}

type pirReq struct {
	reqtype  int
	response chan []byte
}

// Server is a connection and state for a running PIR Server.
type Server struct {
	sock          net.Conn
	responseQueue chan *pirReq
	CellLength    int
	CellCount     int
	BatchSize     int
	DB            *DB
}

const pirCommands = `123`
const defaultSocket = "pir.socket"

// Connect opens a Server for communication with a unix-socket representating a
// running PIR Daemon.
func Connect(socket string) (*Server, error) {
	if len(socket) == 0 {
		socket = defaultSocket
	}

	sock, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}

	server := new(Server)
	server.sock = sock
	server.responseQueue = make(chan *pirReq)

	go server.watchResponses()

	return server, nil
}

func (s *Server) watchResponses() {
	var responseSize int

	for req := range s.responseQueue {
		if req.reqtype == 0 { // reconfigure
			//future buffers will be the new size.
			responseSize = s.CellLength * s.BatchSize
		} else if req.reqtype == -1 { //close
			return
		} else {
			response := make([]byte, responseSize)
			readAmt := 0
			for readAmt < responseSize {
				count, err := s.sock.Read(response[readAmt:])
				readAmt += count
				if count <= 0 || err != nil {
					return
				}
			}
			req.response <- response
		}
	}
}

// Disconnect closes a Server connection
func (s *Server) Disconnect() error {
	s.responseQueue <- &pirReq{-1, nil}
	defer close(s.responseQueue)
	return s.sock.Close()
}

// Configure sets the size of the DB and operational parameters.
func (s *Server) Configure(celllength int, cellcount int, batchsize int) error {
	s.BatchSize = batchsize
	s.CellCount = cellcount
	s.CellLength = celllength

	if s.CellCount%8 != 0 || s.CellLength%8 != 0 {
		return errors.New("invalid sizing of database; everything needs to be multiples of 8 bytes")
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(celllength))
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
	s.responseQueue <- &pirReq{0, nil}
	return nil
}

// GetDB provides direct access to the DB of the Server.
func (s *Server) GetDB() (*DB, error) {
	if s.CellCount == 0 || s.CellLength == 0 {
		return nil, errors.New("pir server unconfigured")
	}
	db := new(DB)
	shmid, err := xusyscall.Shmget(0, s.CellLength*s.CellCount, xusyscall.IPC_CREAT|xusyscall.IPC_EXCL|0777)
	if err != nil {
		return nil, err
	}
	db.shmid = shmid
	db.DB, err = xusyscall.Shmat(db.shmid, false)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// SetDB updates the PIR Server Database
func (s *Server) SetDB(db *DB) error {
	if _, err := s.sock.Write([]byte{pirCommands[2]}); err != nil {
		return err
	}

	fmt.Printf("DB being set for shmid %d\n", db.shmid)
	dbptrarr := make([]byte, 4)
	binary.LittleEndian.PutUint32(dbptrarr, uint32(db.shmid))
	if _, err := s.sock.Write(dbptrarr); err != nil {
		return err
	}
	s.DB = db
	return nil
}

// Free releases memory for a DB instance
func (db *DB) Free() error {
	xusyscall.Shmrm(db.shmid)
	return xusyscall.Shmdt(db.DB)
}

// Read makes a PIR request against the server.
func (s *Server) Read(masks []byte, responseChan chan []byte) error {
	if s.DB == nil || s.CellCount == 0 {
		return errors.New("db not configured")
	}

	if len(masks) != (s.CellCount*s.BatchSize)/8 {
		return errors.New("wrong mask length")
	}

	if _, err := s.sock.Write([]byte{pirCommands[0]}); err != nil {
		return err
	}

	if _, err := s.sock.Write(masks); err != nil {
		return err
	}
	s.responseQueue <- &pirReq{1, responseChan}
	return nil
}
