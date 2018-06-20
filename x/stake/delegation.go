package stake

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Delegation represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
type Delegation struct {
	DelegatorAddr sdk.Address `json:"delegator_addr"`
	ValidatorAddr sdk.Address `json:"validator_addr"`
	Shares        sdk.Rat     `json:"shares"`
	Height        int64       `json:"height"` // Last height bond updated
}

func (b Delegation) equal(b2 Delegation) bool {
	return bytes.Equal(b.DelegatorAddr, b2.DelegatorAddr) &&
		bytes.Equal(b.ValidatorAddr, b2.ValidatorAddr) &&
		b.Height == b2.Height &&
		b.Shares.Equal(b2.Shares)
}

// ensure fulfills the sdk validator types
var _ sdk.Delegation = Delegation{}

// nolint - for sdk.Delegation
func (b Delegation) GetDelegator() sdk.Address { return b.DelegatorAddr }
func (b Delegation) GetValidator() sdk.Address { return b.ValidatorAddr }
func (b Delegation) GetBondShares() sdk.Rat    { return b.Shares }

//Human Friendly pretty printer
func (b Delegation) HumanReadableString() (string, error) {
	bechAcc, err := sdk.Bech32ifyAcc(b.DelegatorAddr)
	if err != nil {
		return "", err
	}
	bechVal, err := sdk.Bech32ifyAcc(b.ValidatorAddr)
	if err != nil {
		return "", err
	}
	resp := "Delegation \n"
	resp += fmt.Sprintf("Delegator: %s\n", bechAcc)
	resp += fmt.Sprintf("Validator: %s\n", bechVal)
	resp += fmt.Sprintf("Shares: %s", b.Shares.String())
	resp += fmt.Sprintf("Height: %d", b.Height)

	return resp, nil
}
