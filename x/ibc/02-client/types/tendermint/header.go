package tendermint

import (
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ exported.Header = Header{}

// Header defines the Tendermint consensus Header
type Header struct {
	// TODO: define Tendermint header type manually, don't use tmtypes
	tmtypes.SignedHeader
	ValidatorSet     *tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
	NextValidatorSet *tmtypes.ValidatorSet `json:"next_validator_set" yaml:"next_validator_set"`
}

// Kind defines that the Header is a Tendermint consensus algorithm
func (header Header) Kind() exported.Kind {
	return exported.Tendermint
}

// GetHeight returns the current height
func (header Header) GetHeight() uint64 {
	return uint64(header.Height)
}

var _ exported.Evidence = Evidence{}

// Evidence defines two disctinct Tendermint headers used to submit a client misbehaviour
// TODO: use evidence module's types
type Evidence struct {
	Header1 Header `json:"header_one" yaml:"header_one"`
	Header2 Header `json:"header_two" yaml:"header_two"`
}

// H1 returns the first header
func (e Evidence) H1() exported.Header {
	return e.Header1
}

// H2 returns the second header
func (e Evidence) H2() exported.Header {
	return e.Header2
}
