package common

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestTrustDomainMarshaling(t *testing.T) {
	tdc := NewTrustDomainConfig("testing", "0.0.0.0", true, false)
	if tdc == nil {
		t.Fatal("Failed to make trust domain.")
	}
	publicBytes, err := json.Marshal(tdc)
	if err != nil {
		t.Fatalf("Failed to serialize Trust Domain.")
	}

	publicDomain := new(TrustDomainConfig)
	err = publicDomain.UnmarshalJSON(publicBytes)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(publicDomain.privateKey[:], tdc.privateKey[:]) == 0 {
		t.Fatal("Serialization should not re-create private key")
	}
	if bytes.Compare(publicDomain.PublicKey[:], tdc.PublicKey[:]) != 0 {
		t.Fatal("Serialization should re-create public key")
	}

	privatebytes, err := json.Marshal(tdc.Private())
	if err != nil {
		t.Fatal(err)
	}

	privateDomain := new(TrustDomainConfig)
	err = privateDomain.UnmarshalJSON(privatebytes)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(privateDomain.privateKey[:], tdc.privateKey[:]) != 0 {
		t.Fatal("Serialization of private() should re-create private key")
	}
}
