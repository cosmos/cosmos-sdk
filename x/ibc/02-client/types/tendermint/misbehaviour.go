package tendermint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"

	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Evidence is a wrapper over tendermint's DuplicateVoteEvidence
// that implements Evidence interface expected by ICS-02
type Evidence struct {
	tmtypes.DuplicateVoteEvidence
	ChainID        string
	ValPubKey      crypto.PubKey
	ValidatorPower int64
	TotalPower     int64
}

// Type implements exported.Evidence interface
func (ev Evidence) Route() string {
	return exported.ClientTypeTendermint
}

// Type implements exported.Evidence interface
func (ev Evidence) Type() string {
	return exported.ClientTypeTendermint
}

// String implements exported.Evidence interface
func (ev Evidence) String() string {
	return ev.DuplicateVoteEvidence.String()
}

// ValidateBasic implements exported.Evidence interface
func (ev Evidence) ValidateBasic() sdk.Error {
	err := ev.DuplicateVoteEvidence.ValidateBasic()
	if err != nil {
		return nil
	}
	return errors.ErrInvalidEvidence(errors.DefaultCodespace, err.Error())
}

// GetConsensusAddress implements exported.Evidence interface
func (ev Evidence) GetConsensusAddress() sdk.ConsAddress {
	return sdk.ConsAddress(ev.DuplicateVoteEvidence.Address())
}

// GetHeight implements exported.Evidence interface
func (ev Evidence) GetHeight() int64 {
	return ev.DuplicateVoteEvidence.Height()
}

// GetValidatorPower implements exported.Evidence interface
func (ev Evidence) GetValidatorPower() int64 {
	return ev.ValidatorPower
}

// GetTotalPower implements exported.Evidence interface
func (ev Evidence) GetTotalPower() int64 {
	return ev.TotalPower
}

// CheckMisbehaviour checks if the evidence provided is a misbehaviour
func CheckMisbehaviour(evidence Evidence) error {
	return evidence.DuplicateVoteEvidence.Verify(evidence.ChainID, evidence.ValPubKey)
}
