package libtalek

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/agl/ed25519"
	"github.com/privacylab/talek/bloom"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"golang.org/x/crypto/nacl/box"
)

// Topic is a writiable Talek log.
// A topic is created by calling NewTopic().
// New items are published with a client via client.Publish(&topic, "Msg").
// Messages can be read from the topic through its contained handle.
type Topic struct {

	// For updates?
	Id uint64

	// For authenticity
	// TODO: this should ratchet.
	SigningPrivateKey *[64]byte `json:",omitempty"`

	Handle
}

// PublishingOverhead represents the number of additional bytes used by encryption and signing.
const PublishingOverhead = box.Overhead + ed25519.SignatureSize

// NewTopic creates a new Topic, or fails if the system randomness isn't
// appropriately configured.
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
	t.Handle.Seed1 = seed1
	t.Handle.Seed2 = seed2
	if err = initHandle(&t.Handle); err != nil {
		return
	}

	// Create shared secret
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)

	t.Handle.SharedSecret = &sharedKey

	// Create signing secrets
	t.Handle.SigningPublicKey, t.SigningPrivateKey, err = ed25519.GenerateKey(rand.Reader)

	return
}

func (t *Topic) GeneratePublish(commonConfig *common.CommonConfig, message []byte) (*common.WriteArgs, error) {
	args := &common.WriteArgs{}
	bucket1, bucket2 := t.Handle.nextBuckets(commonConfig)
	args.Bucket1 = bucket1
	args.Bucket2 = bucket2
	var seqNoBytes [24]byte
	_ = binary.PutUvarint(seqNoBytes[:], t.Seqno)

	t.Handle.Seqno++
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

// @TODO: long-term, keys should ratchet, so that messages outside of the
// active range become refutable. perhaps this could alternatively be done with
// a server managed primative, with releases of a rachet update as each DB epoch
// advances.
func (t *Topic) encrypt(plaintext []byte, nonce *[24]byte) ([]byte, error) {
	buf := make([]byte, 0, len(plaintext)+box.Overhead)
	_ = box.SealAfterPrecomputation(buf, plaintext, nonce, t.Handle.SharedSecret)
	buf = buf[0:cap(buf)]
	digest := ed25519.Sign(t.SigningPrivateKey, buf)
	return append(buf, digest[:]...), nil
}
