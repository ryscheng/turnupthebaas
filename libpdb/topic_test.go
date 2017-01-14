package libpdb

import (
	"crypto/rand"
	"fmt"
	"github.com/ryscheng/pdb/common"
	"strconv"
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	fmt.Printf("TestEncryptDecrypt:\n")
	password := ""
	plaintext := "Hello world"
	nonce := []byte("012345678901")
	th, err := NewTopic(password, 0)
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	ciphertext, err := th.encrypt([]byte(plaintext), nonce)
	if err != nil {
		t.Fatalf("Error encrypting plaintext: %v\n", err)
	}
	result, err := th.Decrypt(ciphertext, nonce)
	if err != nil {
		t.Fatalf("Error decrypting ciphertext: %v\n", err)
	}
	if strings.Compare(plaintext, string(result)) != 0 {
		t.Fatalf("Invalid decrypted value: %v. Expected:%v\n", string(result), plaintext)
	}

	//fmt.Printf("%v", string(result))
	fmt.Printf("... done \n")
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
	password := ""
	th, err := NewTopic(password, 0)
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, err := th.GeneratePublish(config, 1, plaintext)
	if err != nil {
		t.Fatalf("Error creating WriteArgs: %v\n", err)
	}

	fmt.Printf("len(args.Bucket1)=8; len(args.Bucket2)=8; len(args.Data)=%v; len(args.InterestVector)=%v\n", len(args.Data), len(args.InterestVector))

	fmt.Printf("... done \n")
}

func BenchmarkNewTopic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewTopic(strconv.Itoa(i), 0)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	password := ""
	plaintext := make([]byte, 1024, 1024)
	_, err := rand.Read(plaintext)
	if err != nil {
		b.Fatalf("Error creating plaintext: %v\n", err)
	}
	nonce := []byte("012345678901")
	th, err := NewTopic(password, 0)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = th.encrypt(plaintext, nonce)
		if err != nil {
			b.Fatalf("Error encrypting %v: %v\n", i, err)
		}
	}
}

func BenchmarkEncryptDecrypt(b *testing.B) {
	password := ""
	var ciphertext []byte
	plaintext := make([]byte, 1024, 1024)
	_, err := rand.Read(plaintext)
	if err != nil {
		b.Fatalf("Error creating plaintext: %v\n", err)
	}
	nonce := []byte("012345678901")
	th, err := NewTopic(password, 0)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ciphertext, err = th.encrypt(plaintext, nonce)
		if err != nil {
			b.Fatalf("Error encrypting %v: %v\n", i, err)
		}
		_, err = th.Decrypt(ciphertext, nonce)
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
	password := ""
	th, err := NewTopic(password, 0)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = th.GeneratePublish(config, uint64(i), plaintext)
	}
}
