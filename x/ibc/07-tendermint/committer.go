package tendermint

import (
	tmtypes "github.com/tendermint/tendermint/types"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ clientexported.Committer = Committer{}

// Committer definites a Tendermint Committer
type Committer struct {
	*tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
	Height                uint64 `json:"height" yaml:"height"`
	NextValSetHash        []byte `json:"next_valset_hash" yaml:"next_valset_hash"`
}

// ClientType implements exported.Committer interface
func (c Committer) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// GetHeight implements exported.Committer interface
func (c Committer) GetHeight() uint64 {
	return c.Height
}
