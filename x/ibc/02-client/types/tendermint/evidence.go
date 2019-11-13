package tendermint

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
)

var _ evidenceexported.Evidence = Evidence{}

// Evidence is a wrapper over tendermint's DuplicateVoteEvidence
// that implements Evidence interface expected by ICS-02
type Evidence struct {
	*tmtypes.DuplicateVoteEvidence
	ChainID        string `json:"chain_id" yaml:"chain_id"`
	ValidatorPower int64  `json:"val_power" yaml:"val_power"`
	TotalPower     int64  `json:"total_power" yaml:"total_power"`
}

// Route implements Evidence interface
func (ev Evidence) Route() string {
	return "client"
}

// Type implements Evidence interface
func (ev Evidence) Type() string {
	return "client_misbehaviour"
}

// String implements Evidence interface
func (ev Evidence) String() string {
	bz, err := yaml.Marshal(ev)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// Hash implements Evidence interface
func (ev Evidence) Hash() cmn.HexBytes {
	return tmhash.Sum(SubModuleCdc.MustMarshalBinaryBare(ev))
}

// ValidateBasic implements Evidence interface
func (ev Evidence) ValidateBasic() error {
	if ev.DuplicateVoteEvidence == nil {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "duplicate evidence is nil")
	}
	err := ev.DuplicateVoteEvidence.ValidateBasic()
	if err != nil {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, err.Error())
	}
	if ev.ChainID == "" {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "chainID is empty")
	}
	if ev.ValidatorPower <= 0 {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, fmt.Sprintf("invalid Validator Power: %d", ev.ValidatorPower))
	}
	if ev.TotalPower < ev.ValidatorPower {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, fmt.Sprintf("invalid Total Power: %d", ev.TotalPower))
	}
	return nil
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
