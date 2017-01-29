package libpdb

import (
	"crypto/rand"
	"fmt"
	"github.com/ryscheng/pdb/common"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	fmt.Printf("TestEncryptDecrypt:\n")
	plaintext := "Hello world"
	var nonce [24]byte
	copy(nonce[:],[]byte("012345678901"))
	th, err := NewTopic()
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	sub, err := th.CreateSubscription()
	if err != nil {
		t.Fatalf("Failed to derive subscription from topic: %v\n", err)
	}
	ciphertext, err := th.encrypt([]byte(plaintext), &nonce)
	if err != nil {
		t.Fatalf("Error encrypting plaintext: %v\n", err)
	}
	result, err := sub.Decrypt(ciphertext, &nonce)
	if err != nil {
		t.Fatalf("Error decrypting ciphertext: %v, %v\n", ciphertext, err)
	}
	if plaintext != string(result) {
		t.Fatalf("Invalid decrypted value: %v, %v", len(plaintext), len(result))
	}

	//fmt.Printf("%v", string(result))
	fmt.Printf("... done \n")
}

func TestSerializeRestore(t *testing.T) {
	topic, err := NewTopic()
	if err != nil {
		t.Fatalf("Error creating topic: %v\n", err)
	}
	data, err := topic.MarshalBinary()
	if err != nil {
		t.Fatalf("Unable to serialize topic: %v\n", err)
	}
	clone := Topic{}
	err = clone.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("Unable to restore topic: %v\n", err)
	}

	topic.Seqno = 100
	data, err = topic.MarshalBinary()
	err = clone.UnmarshalBinary(data)
	if err != nil || clone.Seqno != 100 {
		t.Fatalf("Sequence number not saved with topic.\n")
	}
}

func TestGeneratePublish(t *testing.T) {
	fmt.Printf("TestGeneratePublish:\n")
	config := &common.CommonConfig{}
	config.NumBuckets = 100
	config.BucketDepth = 2
	config.DataSize = 1024
	config.MaxLoadFactor = 1.0
	config.BloomFalsePositive = 0.1

	plaintext := make([]byte, config.DataSize, config.DataSize)
	_, err := rand.Read(plaintext)
	if err != nil {
		t.Fatalf("Error creating plaintext: %v\n", err)
	}
	th, err := NewTopic()
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, err := th.GeneratePublish(config, plaintext)
	if err != nil {
		t.Fatalf("Error creating WriteArgs: %v\n", err)
	}

	fmt.Printf("len(args.Bucket1)=8; len(args.Bucket2)=8; len(args.Data)=%v; len(args.InterestVector)=%v\n", len(args.Data), len(args.InterestVector))

	fmt.Printf("... done \n")
}

func BenchmarkNewTopic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewTopic()
	}
}

func BenchmarkEncrypt(b *testing.B) {
	plaintext := make([]byte, 1024, 1024)
	_, err := rand.Read(plaintext)
	if err != nil {
		b.Fatalf("Error creating plaintext: %v\n", err)
	}
	var nonce [24]byte
	copy(nonce[:], []byte("012345678901"))
	th, err := NewTopic()
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = th.encrypt(plaintext, &nonce)
		if err != nil {
			b.Fatalf("Error encrypting %v: %v\n", i, err)
		}
	}
}

func BenchmarkEncryptDecrypt(b *testing.B) {
	var ciphertext []byte
	plaintext := make([]byte, 1024, 1024)
	_, err := rand.Read(plaintext)
	if err != nil {
		b.Fatalf("Error creating plaintext: %v\n", err)
	}
	var nonce [24]byte
	copy(nonce[:], []byte("012345678901"))
	th, err := NewTopic()
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	sub, err := th.CreateSubscription()
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ciphertext, err = th.encrypt(plaintext, &nonce)
		if err != nil {
			b.Fatalf("Error encrypting %v: %v\n", i, err)
		}
		_, err = sub.Decrypt(ciphertext, &nonce)
		if err != nil {
			b.Fatalf("Error decrypting %v: %v\n", i, err)
		}
	}
}

func BenchmarkGeneratePublishN10K(b *testing.B) {
	HelperBenchmarkGeneratePublish(b, 100)
}
func BenchmarkGeneratePublishN100K(b *testing.B) {
	HelperBenchmarkGeneratePublish(b, 1000)
}
func BenchmarkGeneratePublishN1M(b *testing.B) {
	HelperBenchmarkGeneratePublish(b, 10000)
}

func HelperBenchmarkGeneratePublish(b *testing.B, BucketDepth int) {
	config := &common.CommonConfig{}
	config.NumBuckets = 100
	config.BucketDepth = BucketDepth
	config.DataSize = 1024
	config.MaxLoadFactor = 1.0
	config.BloomFalsePositive = 0.0001
	plaintext := make([]byte, config.DataSize, config.DataSize)
	_, err := rand.Read(plaintext)
	if err != nil {
		b.Fatalf("Error creating plaintext: %v\n", err)
	}
	th, err := NewTopic()
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = th.GeneratePublish(config, plaintext)
	}
}
