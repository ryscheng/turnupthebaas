package libpdb

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	password := ""
	plaintext := "Hello world"
	fmt.Printf("TestEncryptDecrypt\n")
	th, err := NewTopicHandle(password)
	if err != nil {
		t.Fatalf("Error creating topic handle: %v\n", err)
	}
	ciphertext, nonce, err := th.Encrypt([]byte(plaintext))
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

func BenchmarkNewTopicHandle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewTopicHandle(strconv.Itoa(i))
	}
}

func BenchmarkEncrypt(b *testing.B) {
	password := ""
	th, err := NewTopicHandle(password)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err = th.Encrypt([]byte(strconv.Itoa(i)))
		if err != nil {
			b.Fatalf("Error encrypting %v: %v\n", i, err)
		}
	}
}

func BenchmarkEncryptDecrypt(b *testing.B) {
	password := ""
	th, err := NewTopicHandle(password)
	var ciphertext []byte
	var nonce []byte
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ciphertext, nonce, err = th.Encrypt([]byte(strconv.Itoa(i)))
		if err != nil {
			b.Fatalf("Error encrypting %v: %v\n", i, err)
		}
		_, err = th.Decrypt(ciphertext, nonce)
		if err != nil {
			b.Fatalf("Error decrypting %v: %v\n", i, err)
		}
	}
}
