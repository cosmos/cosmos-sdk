package v038

import (
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// DONTCOVER
// nolint

// Default parameter values
const (
	ModuleName            = "evidence"
	DefaultParamspace     = ModuleName
	DefaultMaxEvidenceAge = 60 * 2 * time.Second
)

// Evidence type constants
const (
	RouteEquivocation = "equivocation"
	TypeEquivocation  = "equivocation"
)

var (
	amino = codec.New()

	// ModuleCdc references the global x/evidence module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/evidence and
	// defined at the application level.
	ModuleCdc = codec.NewHybridCodec(amino, types.NewInterfaceRegistry())
)

// Evidence defines the contract which concrete evidence types of misbehavior
// must implement.
type Evidence interface {
	Route() string
	Type() string
	String() string
	Hash() tmbytes.HexBytes
	ValidateBasic() error

	// Height at which the infraction occurred
	GetHeight() int64
}

// Params defines the total set of parameters for the evidence module
type Params struct {
	MaxEvidenceAge time.Duration `json:"max_evidence_age" yaml:"max_evidence_age"`
}

// GenesisState defines the evidence module's genesis state.
type GenesisState struct {
	Params   Params     `json:"params" yaml:"params"`
	Evidence []Evidence `json:"evidence" yaml:"evidence"`
}

var _ Evidence = &Equivocation{}

// Route returns the Evidence Handler route for an Equivocation type.
func (e *Equivocation) Route() string { return RouteEquivocation }

// Type returns the Evidence Handler type for an Equivocation type.
func (e *Equivocation) Type() string { return TypeEquivocation }

func (e *Equivocation) String() string {
	bz, _ := yaml.Marshal(e)
	return string(bz)
}

// Hash returns the hash of an Equivocation object.
func (e *Equivocation) Hash() tmbytes.HexBytes {
	return tmhash.Sum(ModuleCdc.MustMarshalBinaryBare(e))
}

// ValidateBasic performs basic stateless validation checks on an Equivocation object.
func (e *Equivocation) ValidateBasic() error {
	if e.Time.IsZero() {
		return fmt.Errorf("invalid equivocation time: %s", e.Time)
	}
	if e.Height < 1 {
		return fmt.Errorf("invalid equivocation height: %d", e.Height)
	}
	if e.Power < 1 {
		return fmt.Errorf("invalid equivocation validator power: %d", e.Power)
	}
	if e.ConsensusAddress.Empty() {
		return fmt.Errorf("invalid equivocation validator consensus address: %s", e.ConsensusAddress)
	}

	return nil
}

// GetConsensusAddress returns the validator's consensus address at time of the
// Equivocation infraction.
func (e Equivocation) GetConsensusAddress() sdk.ConsAddress {
	return e.ConsensusAddress
}

// GetHeight returns the height at time of the Equivocation infraction.
func (e Equivocation) GetHeight() int64 {
	return e.Height
}

// GetTime returns the time at time of the Equivocation infraction.
func (e Equivocation) GetTime() time.Time {
	return e.Time
}

// GetValidatorPower returns the validator's power at time of the Equivocation
// infraction.
func (e Equivocation) GetValidatorPower() int64 {
	return e.Power
}

// GetTotalPower is a no-op for the Equivocation type.
func (e Equivocation) GetTotalPower() int64 { return 0 }

// ConvertDuplicateVoteEvidence converts a Tendermint concrete Evidence type to
// SDK Evidence using Equivocation as the concrete type.
func ConvertDuplicateVoteEvidence(dupVote abci.Evidence) exported.Evidence {
	return &Equivocation{
		Height:           dupVote.Height,
		Power:            dupVote.Validator.Power,
		ConsensusAddress: sdk.ConsAddress(dupVote.Validator.Address),
		Time:             dupVote.Time,
	}
}
