package types

import "cosmossdk.io/math"

func (v Validator) GetTokens() math.Int { return v.Tokens }

func NewDescription(moniker, identity, website, securityContact, details string) Description {
	return Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: securityContact,
		Details:         details,
	}
}
