package server

import (
  "bytes"
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/pir"
  "time"
)

import "testing"

func testConf() common.GlobalConfig {
  return common.GlobalConfig{
    512, // num buckets
    4, // depth
    0, //window size?
    512, // data size
    8, // batch size
    0.95, // bloom false positive
    0.95, // max load
    0.02, // load step
    time.Second, //write interval
    time.Second, //read interval
    nil, // trust domains
  }
}

func TestShardSanity(t *testing.T) {
	status := make(chan int)
	go pir.CreateMockServer(status, "pir.socket")
	<-status

  shard := NewShard("Test Shard", testConf())
  if shard == nil {
    t.Error("Failed to create shard.")
    return
  }

  shard.Write(&common.WriteArgs{0, 1, bytes.NewBufferString("Magic").Bytes(), []byte{}, 0}, &common.WriteReply{})

  shard.Table.Flop()

  replychan := make(chan *common.BatchReadReply)
  reqs := make([]common.ReadArgs, 8)

  rv := make([]byte, 512)
  req := common.PirArgs{rv, nil}
  for i := 0; i < 8; i ++ {
    req.RequestVector[i] |= 1
    reqs[i] = common.ReadArgs{[]common.PirArgs{req}}
  }
  shard.BatchRead(&common.BatchReadArgs{reqs, common.Range{0,0, nil}, 0}, replychan)

  reply := <-replychan
  if reply.Replies[0].Data[0] != bytes.NewBufferString("Magic").Bytes()[0] {
    t.Error("Failed to round-trip a write.")
    status <- 1
    return
  }

	status <- 1
}
