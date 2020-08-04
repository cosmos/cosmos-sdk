package v038

import (
	"time"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DONTCOVER
// nolint

// Default parameter values
const (
	ModuleName            = "evidence"
	DefaultParamspace     = ModuleName
	DefaultMaxEvidenceAge = 60 * 2 * time.Second
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

// Equivocation implements the Evidence interface and defines evidence of double
// signing misbehavior.
type Equivocation struct {
	Height           int64           `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Time             time.Time       `protobuf:"bytes,2,opt,name=time,proto3,stdtime" json:"time"`
	Power            int64           `protobuf:"varint,3,opt,name=power,proto3" json:"power,omitempty"`
	ConsensusAddress sdk.ConsAddress `protobuf:"bytes,4,opt,name=consensus_address,json=consensusAddress,proto3,casttype=github.com/cosmos/cosmos-sdk/types.ConsAddress" json:"consensus_address,omitempty" yaml:"consensus_address"`
}

func (m *Equivocation) Reset()      { *m = Equivocation{} }
func (*Equivocation) ProtoMessage() {}
func (*Equivocation) Descriptor() ([]byte, []int) {
	return []byte{}, []int{1}
}
