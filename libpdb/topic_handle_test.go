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
	th, err := NewTopicHandle(password)
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

func TestPublish(t *testing.T) {
	fmt.Printf("TestPublish:\n")
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
	th, err := NewTopicHandle(password)
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, err := th.Publish(globalConfig, 1, plaintext)
	if err != nil {
		t.Fatalf("Error creating WriteArgs: %v\n", err)
	}

	fmt.Printf("len(args.Bucket1)=8; len(args.Bucket2)=8; len(args.Data)=%v; len(args.InterestVector)=%v\n", len(args.Data), len(args.InterestVector))

	fmt.Printf("... done \n")
}

func BenchmarkNewTopicHandle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewTopicHandle(strconv.Itoa(i))
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
	th, err := NewTopicHandle(password)
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
	th, err := NewTopicHandle(password)
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

func BenchmarkPublish(b *testing.B) {
	globalConfig := &common.GlobalConfig{}
	globalConfig.NumBuckets = 100
	globalConfig.WindowSize = 1000000
	globalConfig.DataSize = 1024
	globalConfig.BloomFalsePositive = 0.0001
	plaintext := make([]byte, globalConfig.DataSize, globalConfig.DataSize)
	_, err := rand.Read(plaintext)
	if err != nil {
		b.Fatalf("Error creating plaintext: %v\n", err)
	}
	password := ""
	th, err := NewTopicHandle(password)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = th.Publish(globalConfig, uint64(i), plaintext)
	}
}
