package libtalek

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/agl/ed25519"
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
	ID uint64

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

	t.ID, _ = binary.Uvarint(id[0:8])
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

// GeneratePublish creates a set of write args for writing message as the next
// entry in this topic log.
func (t *Topic) GeneratePublish(commonConfig *common.Config, message []byte) (*common.WriteArgs, error) {
	args := &common.WriteArgs{}
	bucket1, bucket2 := t.Handle.nextBuckets(commonConfig)
	args.Bucket1 = bucket1
	args.Bucket2 = bucket2
	var seqNoBytes [24]byte
	_ = binary.PutUvarint(seqNoBytes[:], t.Seqno)

	args.InterestVector = t.Handle.nextInterestVector()

	t.Handle.Seqno++
	ciphertext, err := t.encrypt(message, &seqNoBytes)
	if err != nil {
		return nil, err
	}
	args.Data = ciphertext

	// @todo - use new bloom/ implementation
	/**
	bloomFilter := bloom.NewWithEstimates(uint(commonConfig.WindowSize()), commonConfig.BloomFalsePositive)
	idBytes := make([]byte, 8, 20)
	_ = binary.PutUvarint(idBytes, t.ID)
	idBytes = append(idBytes, seqNoBytes[:]...)
	bloomFilter.Add(idBytes)
	**/
	//args.InterestVector, _ = bloomFilter.GobEncode()

	return args, nil
}

// @TODO: long-term, keys should ratchet, so that messages outside of the
// active range become refutable. perhaps this could alternatively be done with
// a server managed primitive, with releases of a rachet update as each DB epoch
// advances.
func (t *Topic) encrypt(plaintext []byte, nonce *[24]byte) ([]byte, error) {
	buf := make([]byte, 0, len(plaintext)+box.Overhead)
	_ = box.SealAfterPrecomputation(buf, plaintext, nonce, t.Handle.SharedSecret)
	buf = buf[0:cap(buf)]
	digest := ed25519.Sign(t.SigningPrivateKey, buf)
	return append(buf, digest[:]...), nil
}

// MarshalText is a compact textual representation of a topic
func (t *Topic) MarshalText() ([]byte, error) {
	handle, err := t.Handle.MarshalText()
	if err != nil {
		return nil, err
	}
	txt := fmt.Sprintf("%x.", *t.SigningPrivateKey)
	return append([]byte(txt), handle...), nil
}

// UnmarshalText restores a topic from its compact textual representation
func (t *Topic) UnmarshalText(text []byte) error {
	parts := bytes.SplitN(text, []byte("."), 2)
	if len(parts) != 2 {
		return errors.New("unparsable topic representation")
	}
	t.SigningPrivateKey = new([64]byte)
	var spk []byte
	_, err := fmt.Sscanf(string(parts[0]), "%x", &spk)
	if err != nil {
		return err
	}
	copy(t.SigningPrivateKey[:], spk)
	return t.Handle.UnmarshalText(parts[1])
}
