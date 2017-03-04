package common

import (
	"bytes"
	"crypto/rand"
	"golang.org/x/crypto/nacl/box"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	msg := &ReadArgs{}
	msg.TD = make([]PirArgs, 3)
	config := make([]*TrustDomainConfig, 3)

	for i := 0; i < 3; i++ {
		config[i] = &TrustDomainConfig{}
		msg.TD[i].RequestVector = make([]byte, 32)
		msg.TD[i].PadSeed = make([]byte, 4)
		serverPub, serverPri, _ := box.GenerateKey(rand.Reader)
		copy(config[i].PublicKey[:], serverPub[:])
		copy(config[i].privateKey[:], serverPri[:])
	}

	encodedArgs, err := msg.Encode(config)
	if err != nil || len(encodedArgs.PirArgs) != 3 {
		t.Fatal(err)
	}

	//for each server...
	for i := 0; i < 3; i++ {
		pir, err := encodedArgs.Decode(i, config[i])
		if err != nil {
			t.Fatalf("Error Decoding TD[%d] len(%d): %v\n", i, len(encodedArgs.PirArgs[i]), err)
		}
		if bytes.Compare(pir.PadSeed, msg.TD[i].PadSeed) != 0 ||
			bytes.Compare(pir.RequestVector, msg.TD[i].RequestVector) != 0 {
			t.Fatalf("Decryption not equal.")
		}
	}
}
