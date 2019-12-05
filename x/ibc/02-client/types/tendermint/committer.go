package tendermint

import (
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ exported.Committer = Committer{}

// Committer definites a Tendermint Committer
type Committer struct {
	*tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
	Height                uint64 `json:"height" yaml:"height"`
	NextValSetHash        []byte `json:"next_valset_hash" yaml:"next_valset_hash"`
}

// Implement exported.Committer interface
func (c Committer) ClientType() exported.ClientType {
	return exported.Tendermint
}

// Implement exported.Committer interface
func (c Committer) GetHeight() uint64 {
	return c.Height
}
