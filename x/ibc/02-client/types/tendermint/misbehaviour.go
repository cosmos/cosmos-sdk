package tendermint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"

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

// ValidateBasic implements exported.Evidence interface
func (ev Evidence) ValidateBasic() error {
	return ev.DuplicateVoteEvidence.ValidateBasic()
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
	err := evidence.DuplicateVoteEvidence.Verify(evidence.ChainID, evidence.ValPubKey)
	return sdkerrrors.Wrap(host.ErrInvalidEvidence, err.Error())
}
