package libpdb

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"github.com/agl/ed25519"
	"github.com/dchest/siphash"
	"github.com/ryscheng/pdb/bloom"
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/drbg"
	"golang.org/x/crypto/nacl/box"
)

type Topic struct {

	// For locating log entries
	Id    uint64
	Seed1 drbg.Seed
	Seed2 drbg.Seed

	// For encrypting / decrypting messages
	sharedSecret *[32]byte
	// For authenticity
	// TODO: this should ratchet.
	signingPrivateKey *[64]byte
	signingPublicKey *[32]byte

	// Current log position
	Seqno uint64
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
	t.Seed1 = *seed1
	t.Seed2 = *seed2

	// Create shared secret
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	var sharedKey [32]byte
	box.Precompute(&sharedKey, pub, priv)
	t.sharedSecret = &sharedKey

	// Create signing secrets
	t.signingPublicKey, t.signingPrivateKey, err = ed25519.GenerateKey(rand.Reader)

	return
}

func (t *Topic) CreateSubscription() (*Subscription, error) {
	sub, err := NewSubscription()
	if err != nil {
		return nil, err
	}

	sub.Seqno = t.Seqno
	sub.Seed1 = t.Seed1
	sub.Seed2 = t.Seed2
	sub.SharedSecret = t.sharedSecret
	sub.SigningPublicKey = t.signingPublicKey

	return sub, nil
}

func (t *Topic) GeneratePublish(commonConfig *common.CommonConfig, message []byte) (*common.WriteArgs, error) {
	args := &common.WriteArgs{}
	var seqNoBytes [24]byte
	_ = binary.PutUvarint(seqNoBytes[:], t.Seqno)

	k0, k1 := t.Seed1.KeyUint128()
	args.Bucket1 = siphash.Hash(k0, k1, seqNoBytes[:]) % commonConfig.NumBuckets

	k0, k1 = t.Seed2.KeyUint128()
	args.Bucket2 = siphash.Hash(k0, k1, seqNoBytes[:]) % commonConfig.NumBuckets

	t.Seqno += 1
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
  buf := make([]byte, 0, len(plaintext) + box.Overhead)
	_ = box.SealAfterPrecomputation(buf, plaintext, nonce, t.sharedSecret)
	buf = buf[0:cap(buf)]
	digest := ed25519.Sign(t.signingPrivateKey, buf)
	return append(buf, digest[:]...), nil
}

type binaryTopic struct {
	Id    uint64
	Seed1 []byte
	Seed2 []byte

	// Keys
	SharedSecret  [32]byte
	SigningPrivateKey [64]byte
	SigningPublicKey [32]byte

	// Last seen sequence number
	Seqno uint64
}

/** Implement BinaryMarshaler / BinaryUnmarshaler for serialization **/
func (t *Topic) MarshalBinary() (data []byte, err error) {
	forExport := binaryTopic{t.Id, t.Seed1.Export(), t.Seed2.Export(), *t.sharedSecret, *t.signingPrivateKey, *t.signingPublicKey, t.Seqno}
	var output bytes.Buffer
	enc := gob.NewEncoder(&output)
	err = enc.Encode(forExport)
	if err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (t *Topic) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var forImport binaryTopic
	err := dec.Decode(&forImport)
	if err != nil {
		return err
	}
	t.Id = forImport.Id
	seed1, err := drbg.ImportSeed(forImport.Seed1)
	if err != nil || seed1 == nil {
		return err
	}
	t.Seed1 = *seed1
	seed2, err := drbg.ImportSeed(forImport.Seed2)
	if err != nil || seed2 == nil {
		return err
	}
	t.Seed2 = *seed2
	t.sharedSecret = &forImport.SharedSecret
	t.signingPrivateKey = &forImport.SigningPrivateKey
	t.signingPublicKey = &forImport.SigningPublicKey
	t.Seqno = forImport.Seqno
	return nil
}
