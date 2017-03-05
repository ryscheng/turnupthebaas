package libtalek

import (
	"encoding/binary"
	"errors"
	"github.com/agl/ed25519"
	"github.com/dchest/siphash"
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

func NewSubscription() (s *Subscription, err error) {
	s = &Subscription{}
	s.Updates = make(chan []byte)

	s.drbg, err = drbg.NewHashDrbg(nil)
	return
}

func (s *Subscription) generatePoll(config *ClientConfig, seqNo uint64) (*common.ReadArgs, *common.ReadArgs, error) {
	if s.SharedSecret == nil || s.SigningPublicKey == nil {
		return nil, nil, errors.New("Subscription not fully initialized")
	}

	args := make([]*common.ReadArgs, 2)
	seqNoBytes := make([]byte, 12)
	_ = binary.PutUvarint(seqNoBytes, seqNo)

	num := len(config.TrustDomains)

	args[0] = &common.ReadArgs{}
	args[0].TD = make([]common.PirArgs, num)
	// The first Trust domain is the one with the explicit bucket bit-flip.
	k0, k1 := s.Seed1.KeyUint128()
	bucket1 := siphash.Hash(k0, k1, seqNoBytes) % config.CommonConfig.NumBuckets
	args[0].TD[0].RequestVector = make([]byte, (config.CommonConfig.NumBuckets+7)/8)
	args[0].TD[0].RequestVector[bucket1/8] |= 1 << (bucket1 % 8)
	args[0].TD[0].PadSeed = make([]byte, drbg.SeedLength)
	s.drbg.FillBytes(args[0].TD[0].PadSeed)

	for j := 1; j < num; j++ {
		args[0].TD[j].RequestVector = make([]byte, (config.CommonConfig.NumBuckets+7)/8)
		s.drbg.FillBytes(args[0].TD[j].RequestVector)
		args[0].TD[j].PadSeed = make([]byte, drbg.SeedLength)
		s.drbg.FillBytes(args[0].TD[j].PadSeed)

		for k := 0; k < len(args[0].TD[j].RequestVector); k++ {
			args[0].TD[0].RequestVector[k] ^= args[0].TD[j].RequestVector[k]
		}
	}

	args[1] = &common.ReadArgs{}
	args[1].TD = make([]common.PirArgs, num)
	// The first Trust domain is the one with the explicit bucket bit-flip.
	k0, k1 = s.Seed2.KeyUint128()
	bucket2 := siphash.Hash(k0, k1, seqNoBytes) % config.CommonConfig.NumBuckets
	args[1].TD[0].RequestVector = make([]byte, (config.CommonConfig.NumBuckets+7)/8)
	args[1].TD[0].RequestVector[bucket2/8] |= 1 << (bucket2 % 8)
	args[1].TD[0].PadSeed = make([]byte, drbg.SeedLength)
	s.drbg.FillBytes(args[1].TD[0].PadSeed)

	for j := 1; j < num; j++ {
		args[1].TD[j].RequestVector = make([]byte, (config.CommonConfig.NumBuckets+7)/8)
		s.drbg.FillBytes(args[1].TD[j].RequestVector)
		args[1].TD[j].PadSeed = make([]byte, drbg.SeedLength)
		s.drbg.FillBytes(args[1].TD[j].PadSeed)

		for k := 0; k < len(args[1].TD[j].RequestVector); k++ {
			args[1].TD[0].RequestVector[k] ^= args[1].TD[j].RequestVector[k]
		}
	}

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

	for i := 0; i < len(args.TD); i++ {
		drbg.Overlay(args.TD[i].PadSeed, data)
	}

	return data
}
