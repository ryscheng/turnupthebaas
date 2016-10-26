package libpdb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"github.com/ryscheng/pdb/drbg"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

type TopicHandle struct {
	Id      uint64
	Seed1   drbg.Seed
	Seed2   drbg.Seed
	EncrKey []byte
	// for use with KDF
	salt       []byte
	iterations int
	keyLen     int
}

func NewTopicHandle(password string) (*TopicHandle, error) {
	t := &TopicHandle{}

	// Random values
	salt := make([]byte, 16)
	_, saltErr := rand.Read(salt)
	id := make([]byte, 8)
	_, idErr := rand.Read(id)
	seed1, seed1Err := drbg.NewSeed()
	seed2, seed2Err := drbg.NewSeed()

	// Return errors from crypto.rand
	if idErr != nil {
		return nil, idErr
	}
	if seed1Err != nil {
		return nil, seed1Err
	}
	if seed2Err != nil {
		return nil, seed2Err
	}
	if saltErr != nil {
		return nil, saltErr
	}

	t.salt = salt
	t.iterations = 4096
	t.keyLen = 32

	t.Id, _ = binary.Uvarint(id[0:8])
	t.Seed1 = *seed1
	t.Seed2 = *seed2
	// secret: password
	// public: salt, iterations, keySize
	t.EncrKey = pbkdf2.Key([]byte(password), t.salt, t.iterations, t.keyLen, sha1.New)

	return t, nil
}

//@todo - can we use seqNo as the nonce?
func (t *TopicHandle) Encrypt(plaintext []byte) ([]byte, []byte, error) {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	block, err := aes.NewCipher(t.EncrKey)
	if err != nil {
		return nil, nil, err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	//fmt.Printf("%x\n", ciphertext)
	return ciphertext, nonce, nil
}

func (t *TopicHandle) Decrypt(ciphertext []byte, nonce []byte) ([]byte, error) {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	block, err := aes.NewCipher(t.EncrKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("%s\n", string(plaintext))
	return plaintext, nil
}
