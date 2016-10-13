package common

type TrustDomainConfig struct {
	name          string
	address       string
	isValid       bool
	isDistributed bool
}

func NewTrustDomainConfig(name string, address string, isValid bool, isDistributed bool) *TrustDomainConfig {
	td := &TrustDomainConfig{}
	td.name = name
	td.address = address
	td.isValid = isValid
	td.isDistributed = isDistributed
	return td
}

func (td *TrustDomainConfig) GetName() (string, bool) {
	if td.isValid == false {
		return "", false
	}
	return td.name, td.isValid
}

func (td *TrustDomainConfig) GetAddress() (string, bool) {
	if td.isValid == false {
		return "", false
	}
	return td.address, td.isValid
}

func (td *TrustDomainConfig) IsDistributed() bool {
	return td.isDistributed
}
