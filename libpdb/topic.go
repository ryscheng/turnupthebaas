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

type Topic struct {
	Id      uint64
	Seed1   drbg.Seed
	Seed2   drbg.Seed
	EncrKey []byte
	// for PIR
	drbg *drbg.HashDrbg
	// for use with KDF
	salt       []byte
	iterations int
	keyLen     int
}

func NewTopic(password string) (*Topic, error) {
	t := &Topic{}

	// Random values
	hashDrbg, drbgErr := drbg.NewHashDrbg(nil)
	salt := make([]byte, 16)
	_, saltErr := rand.Read(salt)
	id := make([]byte, 8)
	_, idErr := rand.Read(id)
	seed1, seed1Err := drbg.NewSeed()
	seed2, seed2Err := drbg.NewSeed()

	// Return errors from crypto.rand
	if drbgErr != nil {
		return nil, drbgErr
	}
	if saltErr != nil {
		return nil, saltErr
	}
	if idErr != nil {
		return nil, idErr
	}
	if seed1Err != nil {
		return nil, seed1Err
	}
	if seed2Err != nil {
		return nil, seed2Err
	}

	t.drbg = hashDrbg
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

func (t *Topic) GeneratePublish(globalConfig *common.GlobalConfig, seqNo uint64, message []byte) (*common.WriteArgs, error) {
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

	// @todo - just send the k bit locations
	bloomFilter := bloom.NewWithEstimates(uint(globalConfig.WindowSize()), globalConfig.BloomFalsePositive)
	idBytes := make([]byte, 8, 20)
	_ = binary.PutUvarint(idBytes, t.Id)
	idBytes = append(idBytes, seqNoBytes...)
	bloomFilter.Add(idBytes)
	//args.InterestVector, _ = bloomFilter.GobEncode()

	return args, nil
}

func (t *Topic) generatePoll(globalConfig *common.GlobalConfig, seqNo uint64) (*common.ReadArgs, *common.ReadArgs, error) {
	args := make([]*common.ReadArgs, 2)
	seqNoBytes := make([]byte, 12)
	_ = binary.PutUvarint(seqNoBytes, seqNo)

	args[0] = &common.ReadArgs{}
	args[0].ForTd = make([]common.PirArgs, len(globalConfig.TrustDomains))
	for j := 0; j < len(globalConfig.TrustDomains); j++ {
		args[0].ForTd[j].RequestVector = make([]byte, globalConfig.NumBuckets/8+1)
		t.drbg.FillBytes(args[0].ForTd[j].RequestVector)
		args[0].ForTd[j].PadSeed = make([]byte, drbg.SeedLength)
		t.drbg.FillBytes(args[0].ForTd[j].PadSeed)
	}
	// @todo - XOR this into the last request vector
	//k0, k1 := t.Seed1.KeyUint128()
	//bucket1 := siphash.Hash(k0, k1, seqNoBytes) % globalConfig.NumBuckets

	args[1] = &common.ReadArgs{}
	args[1].ForTd = make([]common.PirArgs, len(globalConfig.TrustDomains))
	for j := 0; j < len(globalConfig.TrustDomains); j++ {
		args[1].ForTd[j].RequestVector = make([]byte, globalConfig.NumBuckets/8+1)
		t.drbg.FillBytes(args[1].ForTd[j].RequestVector)
		args[1].ForTd[j].PadSeed = make([]byte, drbg.SeedLength)
		t.drbg.FillBytes(args[1].ForTd[j].PadSeed)
	}
	// @todo - XOR this into the last request vector
	//k0, k1 = t.Seed2.KeyUint128()
	//bucket2 := siphash.Hash(k0, k1, seqNoBytes) % globalConfig.NumBuckets

	return args[0], args[1], nil
}

func (t *Topic) retrieveResponse(args *common.ReadArgs, reply *common.ReadReply) []byte {
	data := reply.Data

	for i := 0; i < len(args.ForTd); i++ {
		pad := make([]byte, len(data))
		seed, _ := drbg.ImportSeed(args.ForTd[i].PadSeed)
		hashDrbg, _ := drbg.NewHashDrbg(seed)
		hashDrbg.FillBytes(pad)
		for j := 0; j < len(data); j++ {
			data[j] ^= pad[j]
		}
	}
	return data
}

//@todo - can we use seqNo as the nonce?
func (t *Topic) Encrypt(plaintext []byte, nonce []byte) ([]byte, error) {
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

func (t *Topic) Decrypt(ciphertext []byte, nonce []byte) ([]byte, error) {
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
