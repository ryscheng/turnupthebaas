package libtalek

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/agl/ed25519"
	"github.com/dchest/siphash"
	"github.com/privacylab/talek/bloom"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"golang.org/x/crypto/nacl/box"
)

type Topic struct {

	// For updates?
	Id    uint64

	// For authenticity
	// TODO: this should ratchet.
	SigningPrivateKey *[64]byte

	Subscription
}

func NewTopic() (t *Topic, err error) {
	t = &Topic{}

	// Random values
	id := make([]byte, 8)
	if _, err = rand.Read(id); err != nil {
		return
	}
	seed1, err := drbg.NewSeed()
	if err != nil {
		return
	}
	seed2, err := drbg.NewSeed()
	if err != nil {
		return
	}

	t.Id, _ = binary.Uvarint(id[0:8])
	t.Subscription.Seed1 = *seed1
	t.Subscription.Seed2 = *seed2
	t.Subscription.drbg, err = drbg.NewHashDrbg(nil)


	// Create shared secret
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)
	t.Subscription.SharedSecret = &sharedKey

	// Create signing secrets
	t.Subscription.SigningPublicKey, t.SigningPrivateKey, err = ed25519.GenerateKey(rand.Reader)

	return
}

func (t *Topic) GeneratePublish(commonConfig *common.CommonConfig, message []byte) (*common.WriteArgs, error) {
	args := &common.WriteArgs{}
	var seqNoBytes [24]byte
	_ = binary.PutUvarint(seqNoBytes[:], t.Subscription.Seqno)

	k0, k1 := t.Subscription.Seed1.KeyUint128()
	args.Bucket1 = siphash.Hash(k0, k1, seqNoBytes[:]) % commonConfig.NumBuckets

	k0, k1 = t.Subscription.Seed2.KeyUint128()
	args.Bucket2 = siphash.Hash(k0, k1, seqNoBytes[:]) % commonConfig.NumBuckets

	t.Subscription.Seqno += 1
	ciphertext, err := t.encrypt(message, &seqNoBytes)
	if err != nil {
		return nil, err
	}
	args.Data = ciphertext

	// @todo - just send the k bit locations
	bloomFilter := bloom.NewWithEstimates(uint(commonConfig.WindowSize()), commonConfig.BloomFalsePositive)
	idBytes := make([]byte, 8, 20)
	_ = binary.PutUvarint(idBytes, t.Id)
	idBytes = append(idBytes, seqNoBytes[:]...)
	bloomFilter.Add(idBytes)
	//args.InterestVector, _ = bloomFilter.GobEncode()

	return args, nil
}

// @TODO: signing. likely via https://github.com/agl/ed25519
// @TODO: long-term, keys should ratchet, so that messages outside of the
// active range become refutable. perhaps this could alternatively be done with
// a server managed primative, with releases of a rachet update as each DB epoch
// advances.
func (t *Topic) encrypt(plaintext []byte, nonce *[24]byte) ([]byte, error) {
	buf := make([]byte, 0, len(plaintext)+box.Overhead)
	_ = box.SealAfterPrecomputation(buf, plaintext, nonce, t.Subscription.SharedSecret)
	buf = buf[0:cap(buf)]
	digest := ed25519.Sign(t.SigningPrivateKey, buf)
	return append(buf, digest[:]...), nil
}
