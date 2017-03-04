package libtalek

import (
	"encoding/binary"
	"github.com/dchest/siphash"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"testing"
	"time"
)

type mockLeader struct {
	ReceivedWrites chan *common.WriteArgs
}

func (m *mockLeader) GetName() string {
	return "Mock Leader"
}
func (m *mockLeader) Ping(args *common.PingArgs, reply *common.PingReply) error {
	return nil
}
func (m *mockLeader) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	m.ReceivedWrites <- args
	return nil
}
func (m *mockLeader) Read(args *common.EncodedReadArgs, reply *common.ReadReply) error {
	return nil
}
func (m *mockLeader) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	return nil
}

func TestWrite(t *testing.T) {
	config := ClientConfig{
		&common.CommonConfig{64, 4, 1024, 0.05, 0.95, 0.05},
		time.Second,
		time.Second,
		[]*common.TrustDomainConfig{common.NewTrustDomainConfig("TestTrustDomain", "127.0.0.1", true, false)},
	}

	writes := make(chan *common.WriteArgs, 1)
	leader := mockLeader{writes}

	c := NewClient("TestClient", config, &leader)
	if c == nil {
		t.Fatalf("Error creating client")
	}

	handle, _ := NewTopic()

	// Recreate the expected buckets to make sure we're seeing
	// the real write.
	var seqNoBytes [24]byte
	_ = binary.PutUvarint(seqNoBytes[:], handle.Seqno)
	// Clone seed so they advance together.
	seedData, _ := handle.Subscription.Seed1.MarshalBinary()
	seed := drbg.Seed{}
	seed.UnmarshalBinary(seedData)
	k0, k1 := seed.KeyUint128()
	bucket := siphash.Hash(k0, k1, seqNoBytes[:]) % 64

	c.Publish(handle, []byte("hello world"))
	write1 := <-writes
	write2 := <-writes
	c.Kill()
	//Due to thread race, there may be a random write made before
	//the requested publish is queued up.
	if write1.Bucket1 != bucket && write2.Bucket1 != bucket {
		t.Fatalf("Didn't get expected write position.")
	}
}

//TODO: test reading.
