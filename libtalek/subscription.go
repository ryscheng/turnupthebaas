package libtalek

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"github.com/agl/ed25519"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"golang.org/x/crypto/nacl/box"
)

type Subscription struct {
	// for random looking pir requests
	drbg *drbg.HashDrbg

	// For learning log positions
	Seed1 drbg.Seed
	Seed2 drbg.Seed

	// For decrypting messages
	SharedSecret     *[32]byte
	SigningPublicKey *[32]byte

	// Current log position
	Seqno uint64

	// Notifications of new messages
	Updates chan []byte
}

func NewSubscription() (*Subscription, error) {
	s := &Subscription{}
	s.Updates = make(chan []byte)

	hashDrbg, drbgErr := drbg.NewHashDrbg(nil)
	if drbgErr != nil {
		return nil, drbgErr
	}
	s.drbg = hashDrbg

	return s, nil
}

func (s *Subscription) generatePoll(config *ClientConfig, seqNo uint64) (*common.ReadArgs, *common.ReadArgs, error) {
	if s.SharedSecret == nil || s.SigningPublicKey == nil {
		return nil, nil, errors.New("Subscription not fully initialized")
	}

	args := make([]*common.ReadArgs, 2)
	seqNoBytes := make([]byte, 12)
	_ = binary.PutUvarint(seqNoBytes, seqNo)

	args[0] = &common.ReadArgs{}
	args[0].ForTd = make([]common.PirArgs, len(config.TrustDomains))
	for j := 0; j < len(config.TrustDomains); j++ {
		args[0].ForTd[j].RequestVector = make([]byte, config.CommonConfig.NumBuckets/8+1)
		s.drbg.FillBytes(args[0].ForTd[j].RequestVector)
		args[0].ForTd[j].PadSeed = make([]byte, drbg.SeedLength)
		s.drbg.FillBytes(args[0].ForTd[j].PadSeed)
	}
	// @todo - XOR topic info into request?
	//k0, k1 := t.Seed1.KeyUint128()
	//bucket1 := siphash.Hash(k0, k1, seqNoBytes) % globalConfig.NumBuckets

	args[1] = &common.ReadArgs{}
	args[1].ForTd = make([]common.PirArgs, len(config.TrustDomains))
	for j := 0; j < len(config.TrustDomains); j++ {
		args[1].ForTd[j].RequestVector = make([]byte, config.CommonConfig.NumBuckets/8+1)
		s.drbg.FillBytes(args[1].ForTd[j].RequestVector)
		args[1].ForTd[j].PadSeed = make([]byte, drbg.SeedLength)
		s.drbg.FillBytes(args[1].ForTd[j].PadSeed)
	}
	// @todo - XOR sopic info into request?
	//k0, k1 = t.Seed2.KeyUint128()
	//bucket2 := siphash.Hash(k0, k1, seqNoBytes) % globalConfig.NumBuckets

	return args[0], args[1], nil
}

func (s *Subscription) Decrypt(cyphertext []byte, nonce *[24]byte) ([]byte, error) {
	if s.SharedSecret == nil || s.SigningPublicKey == nil {
		return nil, errors.New("Subscription improperly initialized")
	}
	cypherlen := len(cyphertext)
	if cypherlen < ed25519.SignatureSize {
		return nil, errors.New("Invalid cyphertext")
	}

	//verify signature
	message := cyphertext[0 : cypherlen-ed25519.SignatureSize]
	var sig [ed25519.SignatureSize]byte
	copy(sig[:], cyphertext[cypherlen-ed25519.SignatureSize:])
	if !ed25519.Verify(s.SigningPublicKey, message, &sig) {
		return nil, errors.New("Invalid Signature")
	}

	//decrypt
	plaintext := make([]byte, 0, cypherlen-box.Overhead-ed25519.SignatureSize)
	_, ok := box.OpenAfterPrecomputation(plaintext, message, nonce, s.SharedSecret)
	if !ok {
		return nil, errors.New("Failed to decrypt.")
	}
	return plaintext[0:cap(plaintext)], nil
}

func (s *Subscription) OnResponse(args *common.ReadArgs, reply *common.ReadReply) {
	msg := s.retrieveResponse(args, reply)
	if msg != nil && s.Updates != nil {
		s.Updates <- msg
	}
}

// TODO: checksum msgs at topic level so if something random comes back it is filtered out.
func (s *Subscription) retrieveResponse(args *common.ReadArgs, reply *common.ReadReply) []byte {
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

func (s *Subscription) MarshalBinary() (data []byte, err error) {
	var output bytes.Buffer
	enc := gob.NewEncoder(&output)
	err = enc.Encode(s)
	if err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (s *Subscription) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(s)
	if err != nil {
		return err
	}
	return nil
}
