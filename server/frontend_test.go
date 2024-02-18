package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/privacylab/talek/common"
)

type mockReplica struct {
	calls []string
}

func (m *mockReplica) Write(args *common.ReplicaWriteArgs, reply *common.ReplicaWriteReply) error {
	if m.calls == nil {
		m.calls = make([]string, 1)
	}
	m.calls = append(m.calls, "write-"+fmt.Sprintf("%d", args.GlobalSeqNo))
	return nil
}
func (m *mockReplica) BatchRead(args *common.BatchReadRequest, reply *common.BatchReadReply) error {
	if m.calls == nil {
		m.calls = make([]string, 1)
	}
	m.calls = append(m.calls, "read-"+fmt.Sprintf("%d", (len(args.Args))))
	reply.Replies = make([]common.ReadReply, len(args.Args))
	return nil
}
func (m *mockReplica) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	return nil
}

func TestFrontendWrite(t *testing.T) {
	back := new(mockReplica)
	serverConfig := &Config{
		WriteInterval: time.Millisecond * 100,
		ReadInterval:  time.Minute,
	}

	f := NewFrontend("testing", serverConfig, []common.ReplicaInterface{back})

	if len(back.calls) != 0 {
		t.Fatalf("there should be no replica calls on startup")
	}

	args := &common.WriteArgs{}
	reply := &common.WriteReply{}
	if err := f.Write(args, reply); err != nil {
		t.Fatal(err)
	}

	l := len(back.calls)
	if l < 1 {
		t.Fatalf("replica should have been written to (%d calls)", len(back.calls))
	}

	time.Sleep(time.Millisecond * 150)

	if len(back.calls) == l {
		t.Fatalf("periodic writes should be occuring.")
	}

	f.Close()
}

func TestFrontendRead(t *testing.T) {
	back := new(mockReplica)
	serverConfig := &Config{
		Config:        &common.Config{},
		ReadInterval:  time.Millisecond * 100,
		WriteInterval: time.Minute,
	}

	f := NewFrontend("testing", serverConfig, []common.ReplicaInterface{back})

	args := &common.EncodedReadArgs{}
	reply := &common.ReadReply{}
	go f.Read(args, reply)

	if len(back.calls) != 0 {
		t.Fatalf("reads should be batched. not immediately sent.")
	}

	time.Sleep(time.Millisecond * 150)

	if len(back.calls) == 0 {
		t.Fatalf("periodic reads should be occuring.")
	}

	f.Close()
}
