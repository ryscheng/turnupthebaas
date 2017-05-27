package libtalek

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"testing"

	"github.com/privacylab/talek/common"
)

func TestEncryptDecrypt(t *testing.T) {
	fmt.Printf("TestEncryptDecrypt:\n")
	plaintext := "Hello world"
	var nonce [24]byte
	nonceint := uint64(12345678)
	_ = binary.PutUvarint(nonce[:], nonceint)
	th, err := NewTopic()
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	var h Handle
	h = th.Handle
	if err != nil {
		t.Fatalf("Failed to derive handle from topic: %v\n", err)
	}
	ciphertext, err := th.encrypt([]byte(plaintext), &nonce)
	if err != nil {
		t.Fatalf("Error encrypting plaintext: %v\n", err)
	}
	h.Seqno = nonceint
	result, err := h.Decrypt(ciphertext, &nonce)
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

	// test binary encoder
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	err = enc.Encode(topic)
	if err != nil {
		t.Fatalf("Unable to serialize topic: %v\n", err)
	}
	dec := gob.NewDecoder(&network)
	clone := Topic{}
	err = dec.Decode(&clone)
	if err != nil {
		t.Fatalf("Unable to restore topic: %v\n", err)
	}

	// test text encoder
	txt, err := topic.MarshalText()
	if err != nil {
		t.Fatalf("Error serializing: %v\n", err)
	}
	fmt.Printf("Serialized topic looks like %s\n", txt)

	clone = Topic{}
	err = clone.UnmarshalText(txt)
	if err != nil {
		t.Fatalf("Could not deserialize: %v\n", err)
	}
	if !bytes.Equal(topic.SigningPrivateKey[:], clone.SigningPrivateKey[:]) || !Equal(&topic.Handle, &clone.Handle) {
		t.Fatalf("serialization lost info!")
	}
}

func TestGeneratePublish(t *testing.T) {
	fmt.Printf("TestGeneratePublish:\n")
	config := &common.Config{}
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
	h := th.Handle
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ciphertext, err = th.encrypt(plaintext, &nonce)
		if err != nil {
			b.Fatalf("Error encrypting %v: %v\n", i, err)
		}
		_, err = h.Decrypt(ciphertext, &nonce)
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
	config := &common.Config{}
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
