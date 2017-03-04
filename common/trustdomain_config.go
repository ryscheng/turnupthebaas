package common

import (
	"crypto/rand"
	"encoding/json"
	"golang.org/x/crypto/nacl/box"
)

type TrustDomainConfig struct {
	Name          string
	Address       string
	IsValid       bool
	IsDistributed bool
	PublicKey     [32]byte
	privateKey    [32]byte
}

type PrivateTrustDomainConfig struct {
	*TrustDomainConfig
	PrivateKey [32]byte
}

// Create a new TrustDomainConfig with a freshly generated keypair.
func NewTrustDomainConfig(name string, address string, isValid bool, isDistributed bool) *TrustDomainConfig {
	td := &TrustDomainConfig{}
	td.Name = name
	td.Address = address
	td.IsValid = isValid
	td.IsDistributed = isDistributed
	pubKey, priKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		td.IsValid = false
		return td
	}
	copy(td.PublicKey[:], pubKey[:])
	copy(td.privateKey[:], priKey[:])
	return td
}

// Marshaled JSON for a TrustDomainConfig may include a 'PrivateKey' when it
// represents the config for a server / trust domain. The custom UnmarshalJSON
// function supports restoring that key.
func (t *TrustDomainConfig) UnmarshalJSON(marshaled []byte) error {
	if len(marshaled) == 0 {
		return nil
	}
	// The union type between TrustDomainCOnfig and PrivateTrustDomainConfig.
	type Config struct {
		PublicKey     [32]byte
		PrivateKey    [32]byte
		Name          string
		Address       string
		IsValid       bool
		IsDistributed bool
	}
	var config Config
	if err := json.Unmarshal(marshaled, &config); err != nil {
		return err
	}

	copy(t.privateKey[:], config.PrivateKey[:])
	copy(t.PublicKey[:], config.PublicKey[:])
	t.Name = config.Name
	t.Address = config.Address
	t.IsValid = config.IsValid
	t.IsDistributed = config.IsDistributed

	return nil
}


// Expose the Private key of a trust domain config for marshalling.
//   bytes, err := json.Marshal(trustdomainconfig.Private())
func (td *TrustDomainConfig) Private() *PrivateTrustDomainConfig {
	PTDC := new(PrivateTrustDomainConfig)
	PTDC.TrustDomainConfig = td
	copy(PTDC.PrivateKey[:], td.privateKey[:])
	return PTDC
}

func (td *TrustDomainConfig) GetName() (string, bool) {
	if td.IsValid == false {
		return "", false
	}
	return td.Name, td.IsValid
}

func (td *TrustDomainConfig) GetAddress() (string, bool) {
	if td.IsValid == false {
		return "", false
	}
	return td.Address, td.IsValid
}
