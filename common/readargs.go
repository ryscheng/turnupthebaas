package common

/**
 * Underlying Algorithm for the client to encrypt read args to the different
 * servers, and servers to decrypt and read the encoded args.
 */

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"errors"

	"golang.org/x/crypto/nacl/box"
)

// Encode encrypts a read request for a given trust domain configuration.
func (r *ReadArgs) Encode(trustDomains []*TrustDomainConfig) (out EncodedReadArgs, err error) {
	pubKey, priKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	nonce := make([]byte, 24)
	_, err = rand.Read(nonce)
	if err != nil {
		return
	}

	// Allocate memory
	copy(out.ClientKey[:], pubKey[:])
	copy(out.Nonce[:], nonce)

	out.PirArgs = make([][]byte, len(trustDomains))

	var msg bytes.Buffer
	enc := gob.NewEncoder(&msg)
	for i := 0; i < len(trustDomains); i++ {
		err = enc.Encode(r.TD[i])
		if err != nil {
			return
		}
		msgBytes := msg.Bytes()
		out.PirArgs[i] = make([]byte, 0, len(msgBytes)+box.Overhead)
		ret := box.Seal(out.PirArgs[i], msgBytes, &out.Nonce, &trustDomains[i].PublicKey, priKey)
		out.PirArgs[i] = ret
	}
	return
}

// Bucket returns the bucket index that a read requests, or -1 for invalid args.
func (r *ReadArgs) Bucket() int {
	finalvec := make([]byte, len(r.TD[0].RequestVector))
	for i := 0; i < len(r.TD); i++ {
		for j := 0; j < len(finalvec); j++ {
			finalvec[j] ^= r.TD[i].RequestVector[j]
		}
	}

	//confirm that there's only 1 bit set, and find it's index.
	bit := -1
	for i := 0; i < 8*len(finalvec); i++ {
		if finalvec[i/8]&(1<<uint(i%8)) != 0 {
			if bit > -1 {
				return -1
			}
			bit = i
		}
	}

	return bit
}

// Decode decrypts a specific trust domain of encoded args to recover the pad and request vector.
func (r *EncodedReadArgs) Decode(id int, trustDomain *TrustDomainConfig) (out PirArgs, err error) {
	if len(r.PirArgs[id]) < box.Overhead {
		err = errors.New("Attempted Decoding of invalid Trust Domain")
		return
	}
	msg := make([]byte, 0, len(r.PirArgs[id])-box.Overhead)
	decrypted, _ := box.Open(msg, r.PirArgs[id], &r.Nonce, &r.ClientKey, &trustDomain.privateKey)
	dec := gob.NewDecoder(bytes.NewBuffer(decrypted))
	err = dec.Decode(&out)
	return
}
