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
	th, err := NewTopic(password)
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	ciphertext, err := th.Encrypt([]byte(plaintext), nonce)
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
	globalConfig := &common.GlobalConfig{}
	globalConfig.NumBuckets = 100
	globalConfig.WindowSize = 10000
	globalConfig.DataSize = 1024
	globalConfig.BloomFalsePositive = 0.1
	//globalConfig.BloomFalsePositive = 0.0001
	plaintext := make([]byte, globalConfig.DataSize, globalConfig.DataSize)
	_, err := rand.Read(plaintext)
	if err != nil {
		t.Fatalf("Error creating plaintext: %v\n", err)
	}
	password := ""
	th, err := NewTopic(password)
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, err := th.GeneratePublish(globalConfig, 1, plaintext)
	if err != nil {
		t.Fatalf("Error creating WriteArgs: %v\n", err)
	}

	fmt.Printf("len(args.Bucket1)=8; len(args.Bucket2)=8; len(args.Data)=%v; len(args.InterestVector)=%v\n", len(args.Data), len(args.InterestVector))

	fmt.Printf("... done \n")
}

func TestGeneratePoll(t *testing.T) {
	fmt.Printf("TestGeneratePoll:\n")
	globalConfig := &common.GlobalConfig{}
	globalConfig.NumBuckets = 1000000
	globalConfig.TrustDomains = make([]*common.TrustDomainConfig, 3)
	//globalConfig.BloomFalsePositive = 0.0001
	password := ""
	th, err := NewTopic(password)
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	args0, _, err := th.generatePoll(globalConfig, 1)
	if err != nil {
		t.Fatalf("Error creating ReadArgs: %v\n", err)
	}

	fmt.Printf("len(args0)=%v; \n", 3*(len(args0.ForTd[0].RequestVector)+len(args0.ForTd[0].PadSeed)))

	fmt.Printf("... done \n")
}

func BenchmarkNewTopic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewTopic(strconv.Itoa(i))
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
	th, err := NewTopic(password)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = th.Encrypt(plaintext, nonce)
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
	th, err := NewTopic(password)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ciphertext, err = th.Encrypt(plaintext, nonce)
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
	HelperBenchmarkGeneratePublish(b, 10000)
}
func BenchmarkGeneratePublishN100K(b *testing.B) {
	HelperBenchmarkGeneratePublish(b, 100000)
}
func BenchmarkGeneratePublishN1M(b *testing.B) {
	HelperBenchmarkGeneratePublish(b, 1000000)
}

func HelperBenchmarkGeneratePublish(b *testing.B, WindowSize int) {
	globalConfig := &common.GlobalConfig{}
	globalConfig.NumBuckets = 100
	globalConfig.WindowSize = WindowSize
	globalConfig.DataSize = 1024
	globalConfig.BloomFalsePositive = 0.0001
	plaintext := make([]byte, globalConfig.DataSize, globalConfig.DataSize)
	_, err := rand.Read(plaintext)
	if err != nil {
		b.Fatalf("Error creating plaintext: %v\n", err)
	}
	password := ""
	th, err := NewTopic(password)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = th.generatePublish(globalConfig, uint64(i), plaintext)
	}
}

func BenchmarkGeneratePollN10K(b *testing.B) {
	HelperBenchmarkGeneratePoll(b, 10000/4)
}
func BenchmarkGeneratePollN100K(b *testing.B) {
	HelperBenchmarkGeneratePoll(b, 100000/4)
}
func BenchmarkGeneratePollN1M(b *testing.B) {
	HelperBenchmarkGeneratePoll(b, 1000000/4)
}

func HelperBenchmarkGeneratePoll(b *testing.B, WindowSize uint64) {
	globalConfig := &common.GlobalConfig{}
	globalConfig.TrustDomains = make([]*common.TrustDomainConfig, 3)
	globalConfig.NumBuckets = WindowSize

	password := ""
	th, err := NewTopic(password)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = th.generatePoll(globalConfig, uint64(i))
	}

}

func BenchmarkRetrieveResponse(b *testing.B) {
	globalConfig := &common.GlobalConfig{}
	globalConfig.TrustDomains = make([]*common.TrustDomainConfig, 3)
	globalConfig.NumBuckets = 10

	password := ""
	th, err := NewTopic(password)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, _, err := th.generatePoll(globalConfig, 1)
	if err != nil {
		b.Fatalf("Error creating ReadArgs: %v\n", err)
	}
	reply := &common.ReadReply{}
	reply.Data = make([]byte, 1024)
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = th.retrieveResponse(args, reply)
	}

}
