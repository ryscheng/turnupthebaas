package common

type TrustDomainConfig struct {
	name    string
	address string
	isValid bool
}

func NewTrustDomainConfig(name string, address string, isValid bool) *TrustDomainConfig {
	td := &TrustDomainConfig{}
	td.name = name
	td.address = address
	td.isValid = isValid
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
