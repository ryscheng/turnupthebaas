package common

type TrustDomainConfig struct {
	Address string
	IsValid bool
}

func NewTrustDomainConfig(address string, isValid bool) *TrustDomainConfig {
	td := &TrustDomainConfig{}
	td.Address = address
	td.IsValid = isValid
	return td
}

func (td *TrustDomainConfig) GetAddress() (string, bool) {
	if td.IsValid == false {
		return "", false
	}
	return td.Address, td.IsValid
}
