package pir

import "bytes"
import "encoding/binary"
import "errors"
import "net"
import "github.com/YoshikiShibata/xusyscall"

type PirDB struct {
	DB    []byte
	shmid int
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

func Connect(socket string) (*PirServer, error) {
	if len(socket) == 0 {
		socket = defaultSocket
	}

	sock, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}

	server := new(PirServer)
	server.sock = sock
	return server, nil
}

func (s *PirServer) Disconnect() error {
	return s.sock.Close()
}

func (s *PirServer) Configure(celllength int, cellcount int, batchsize int) error {
	s.BatchSize = batchsize
	s.CellCount = cellcount
	s.CellLength = celllength

	if s.CellCount %8 != 0 || s.CellLength % 8 != 0 {
		return errors.New("Invalid sizing of database. everything needs to be multiples of 8 bytes.")
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
	return nil
}

func (s *PirServer) GetDB() (*PirDB, error) {
	if s.CellCount == 0 || s.CellLength == 0 {
		return nil, errors.New("PIR Server unconfigured.")
	}
	db := new(PirDB)
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

func (s *PirServer) SetDB(db *PirDB) error {
	if _, err := s.sock.Write([]byte{pirCommands[2]}); err != nil {
		return err
	}

	dbptrarr := make([]byte, 4)
	binary.LittleEndian.PutUint32(dbptrarr, uint32(db.shmid))
	if _, err := s.sock.Write(dbptrarr); err != nil {
		return err
	}
	s.DB = db
	if _, err := s.sock.Read(dbptrarr[0:2]); err != nil {
		return err
	}
	return nil
}

func (db *PirDB) Free() error {
	return xusyscall.Shmdt(db.DB)
}

func (s *PirServer) Read(masks []byte) ([]byte, error) {
	if s.DB == nil || s.CellCount == 0 {
		return nil, errors.New("DB not configured.")
	}

	if len(masks) != (s.CellCount*s.BatchSize)/8 {
		return nil, errors.New("Wrong Mask Length.")
	}

	if _, err := s.sock.Write([]byte{pirCommands[0]}); err != nil {
		return nil, err
	}

	if _, err := s.sock.Write(masks); err != nil {
		return nil, err
	}

	responses := make([]byte, s.CellLength*s.BatchSize)
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
