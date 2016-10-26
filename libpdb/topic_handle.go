package libpdb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"github.com/dchest/siphash"
	"github.com/ryscheng/pdb/bloom"
	"github.com/ryscheng/pdb/common"
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

func (t *TopicHandle) Publish(globalConfig *common.GlobalConfig, seqNo uint64, message []byte) (*common.WriteArgs, error) {
	args := &common.WriteArgs{}
	seqNoBytes := make([]byte, 12)
	_ = binary.PutUvarint(seqNoBytes, seqNo)

	k0, k1 := t.Seed1.KeyUint128()
	args.Bucket1 = siphash.Hash(k0, k1, seqNoBytes) % globalConfig.NumBuckets

	k0, k1 = t.Seed2.KeyUint128()
	args.Bucket2 = siphash.Hash(k0, k1, seqNoBytes) % globalConfig.NumBuckets

	ciphertext, err := t.Encrypt(message, seqNoBytes)
	if err != nil {
		return nil, err
	}
	args.Data = ciphertext

	bloomFilter := bloom.NewWithEstimates(uint(globalConfig.WindowSize), globalConfig.BloomFalsePositive)
	idBytes := make([]byte, 8, 20)
	_ = binary.PutUvarint(idBytes, t.Id)
	idBytes = append(idBytes, seqNoBytes...)
	bloomFilter.Add(idBytes)
	args.InterestVector, _ = bloomFilter.GobEncode()

	return args, nil
}

func (t *TopicHandle) Poll(globalConfig *common.GlobalConfig, seqNo uint64) (*common.ReadArgs, error) {
	return nil, nil
}

//@todo - can we use seqNo as the nonce?
func (t *TopicHandle) Encrypt(plaintext []byte, nonce []byte) ([]byte, error) {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	block, err := aes.NewCipher(t.EncrKey)
	if err != nil {
		return nil, err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	/**
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	**/

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	//fmt.Printf("%x\n", ciphertext)
	return ciphertext, nil
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
