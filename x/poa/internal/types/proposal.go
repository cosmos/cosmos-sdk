package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/crypto"
)

const (
	ProposeCreateValidator = "NewValidatorCreatation"
	ProposeIncreaseWeight  = "NewValidatorWeight"
)

var _ govtypes.Content = MsgProposeCreateValidator{}

type MsgProposeCreateValidator struct {
	Title       string             `json:"title" yaml:"title"`             // title of the validator
	Description string             `json:"description" yaml:"description"` // description of validator
	Validator   MsgCreateValidator `json:"validator" yaml:"validator"`
}

func NewMsgProposeCreateValidator(title, description string, valAddr sdk.ValAddress,
	pubKey crypto.PubKey, valDescription stakingtypes.Description) MsgProposeCreateValidator {
	newValidator := NewMsgCreateValidator(valAddr, pubKey, valDescription)
	return MsgProposeCreateValidator{
		Title:       title,
		Description: description,
		Validator:   newValidator,
	}
}

// GetTitle returns the title of the validator.
func (mpc MsgProposeCreateValidator) GetTitle() string { return mpc.Title }

// GetDescription returns the description of a poa change proposal.
func (mpc MsgProposeCreateValidator) GetDescription() string { return mpc.Description }

// ProposalRoute returns the routing key of a poa change proposal.
func (mpc MsgProposeCreateValidator) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a poa change proposal.
func (mpc MsgProposeCreateValidator) ProposalType() string { return ProposeCreateValidator }

// String implements the stringer interface
func (mpc MsgProposeCreateValidator) String() string {
	des := mpc.Validator.Description
	d := fmt.Sprintf(`
		Title: %s,
		Moinker: %s,
		Identity: %s,
		Website: %s, 
		SecruityContract: %s, 
		Details: %s,
		ValidatorAddress: %s,
		PubKey: %s
		`, mpc.Title, des.Moniker, des.Identity, des.Website,
		des.SecurityContact, des.Details, mpc.Validator.ValidatorAddress.String(), mpc.Validator.PubKey.Address().String())
	return d
}

// ValidateBasic validates the Creation of a validator proposal
func (mpc MsgProposeCreateValidator) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodeSpace, mpc)
	if err != nil {
		return err
	}

	return ValidateChanges(mpc)
}

func ValidateChanges(cVA MsgProposeCreateValidator) sdk.Error {
	if len(cVA.Title) == 0 {
		return params.ErrEmptyChanges(DefaultCodeSpace)
	}
	if cVA.Validator.Description == (stakingtypes.Description{}) {
		return sdk.NewError(stakingtypes.DefaultCodespace, stakingtypes.CodeInvalidInput, "description must be included")
	}
	if cVA.Validator.ValidatorAddress.Empty() {
		return stakingtypes.ErrNilValidatorAddr(DefaultCodeSpace)
	}
	return nil
}

// --------------------------------

type MsgProposeIncreaseWeight struct {
	Title       string                  `json:"title" yaml:"title"`             // title of the validator
	Description string                  `json:"description" yaml:"description"` // description of validator
	Validator   ValidatorIncreaseWeight `json:"validator" yaml:"validator"`
}

func NewMsgProposeIncreaseWeight(t, d string, v ValidatorIncreaseWeight) MsgProposeIncreaseWeight {
	return MsgProposeIncreaseWeight{
		Title:       t,
		Description: d,
		Validator:   v,
	}
}

// GetTitle returns the title of the validator.
func (mpi MsgProposeIncreaseWeight) GetTitle() string { return mpi.Title }

// GetDescription returns the description of a poa change proposal.
func (mpi MsgProposeIncreaseWeight) GetDescription() string { return mpi.Description }

// ProposalRoute returns the routing key of a poa change proposal.
func (mpi MsgProposeIncreaseWeight) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a poa change proposal.
func (mpi MsgProposeIncreaseWeight) ProposalType() string { return ProposeIncreaseWeight }

// String implements the stringer interface
func (mpi MsgProposeIncreaseWeight) String() string {
	val := mpi.Validator
	d := fmt.Sprintf(`
		ValidatorAddress: %s,
		New Weight: %v
		`,
		val.ValidatorAddress.String(), val.NewWeight)
	return d
}

// ValidateBasic validates the Creation of a validator proposal
func (mpi MsgProposeIncreaseWeight) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodeSpace, mpi)
	if err != nil {
		return err
	}

	if len(mpi.Title) == 0 || len(mpi.Description) == 0 {
		return params.ErrEmptyChanges(DefaultCodeSpace)
	}
	if mpi.Validator.ValidatorAddress.Empty() {
		return stakingtypes.ErrNilValidatorAddr(DefaultCodeSpace)
	}
	return nil
}

type ValidatorIncreaseWeight struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	NewWeight        sdk.Int        `json:"new_weight" yaml:"new_weight"`
}

func ValidatorWeightIncrease(va sdk.ValAddress, pb crypto.PubKey, nw sdk.Int) ValidatorIncreaseWeight {
	return ValidatorIncreaseWeight{
		ValidatorAddress: va,
		NewWeight:        nw,
	}
}
