package tendermint

import (
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ exported.Header = Header{}

// Header defines the Tendermint consensus Header
type Header struct {
	tmtypes.SignedHeader
	ValidatorSet     *tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
	NextValidatorSet *tmtypes.ValidatorSet `json:"next_validator_set" yaml:"next_validator_set"`
}

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() exported.ClientType {
	return exported.Tendermint
}

// GetHeight returns the current height
//
// NOTE: also referred as `sequence`
func (h Header) GetHeight() uint64 {
	return uint64(h.Height)
}
