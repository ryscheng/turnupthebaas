package libpdb

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"github.com/ryscheng/pdb/drbg"
	"golang.org/x/crypto/pbkdf2"
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
