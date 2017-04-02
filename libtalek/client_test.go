package libtalek

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/privacylab/talek/common"
)

type mockLeader struct {
	ReceivedWrites chan *common.WriteArgs
	ReceivedReads  chan *common.EncodedReadArgs
}

func (m *mockLeader) GetName(_ *interface{}, reply *string) error {
	*reply = "Mock Leader"
	return nil
}
func (m *mockLeader) GetConfig(_ *interface{}, reply *common.Config) error {
	return nil
}
func (m *mockLeader) Ping(args *common.PingArgs, reply *common.PingReply) error {
	return nil
}
func (m *mockLeader) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	if m.ReceivedWrites != nil {
		m.ReceivedWrites <- args
	}
	return nil
}
func (m *mockLeader) Read(args *common.EncodedReadArgs, reply *common.ReadReply) error {
	if m.ReceivedReads != nil {
		m.ReceivedReads <- args
	}
	return nil
}
func (m *mockLeader) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	return nil
}

func TestWrite(t *testing.T) {
	config := ClientConfig{
		&common.Config{NumBuckets: 64, BucketDepth: 4, DataSize: 1024, BloomFalsePositive: 0.05, MaxLoadFactor: 0.95, LoadFactorStep: 0.05},
		time.Second,
		time.Second,
		[]*common.TrustDomainConfig{common.NewTrustDomainConfig("TestTrustDomain", "127.0.0.1", true, false)},
	}

	writes := make(chan *common.WriteArgs, 1)
	leader := mockLeader{writes, nil}

	c := NewClient("TestClient", config, &leader)
	if c == nil {
		t.Fatalf("Error creating client")
	}

	handle, _ := NewTopic()

	// Recreate the expected buckets to make sure we're seeing
	// the real write.
	bucket, _ := handle.Handle.nextBuckets(config.Config)

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

func TestRead(t *testing.T) {
	config := ClientConfig{
		&common.Config{NumBuckets: 64, BucketDepth: 4, DataSize: 1024, BloomFalsePositive: 0.05, MaxLoadFactor: 0.95, LoadFactorStep: 0.05},
		time.Second,
		time.Second,
		[]*common.TrustDomainConfig{common.NewTrustDomainConfig("TestTrustDomain", "127.0.0.1", true, false)},
	}

	reads := make(chan *common.EncodedReadArgs, 1)
	leader := mockLeader{nil, reads}

	c := NewClient("TestRead", config, &leader)
	if c == nil {
		t.Fatalf("Error creating client")
	}

	handle, _ := NewTopic()

	// Recreate the expected buckets to make sure we're seeing
	// the real write.
	var seqNoBytes [24]byte
	_ = binary.PutUvarint(seqNoBytes[:], handle.Seqno)
	// Clone seed so they advance together.
	bucket, _ := handle.Handle.nextBuckets(config.Config)

	c.Poll(&handle.Handle)
	read1 := <-reads
	read2 := <-reads
	// There may be a random read occuring before the enqueued one.
	c.Kill()

	//Due to thread race, there may be a random read made before
	//the requested poll is queued up.
	decRead1, _ := read1.Decode(0, config.TrustDomains[0])
	decRead2, err := read2.Decode(0, config.TrustDomains[0])
	if err != nil {
		t.Fatalf("Failed to decode read %v", err)
	}
	if decRead1.RequestVector[bucket/8]&(1<<(bucket%8)) == 0 &&
		decRead2.RequestVector[bucket/8]&(1<<(bucket%8)) == 0 {
		t.Fatalf("Read wasn't for the enqueued subscription. %v / %v / %d", decRead1.RequestVector, decRead2.RequestVector, bucket)
	}
}
