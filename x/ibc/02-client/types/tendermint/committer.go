package tendermint

import (
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ exported.Committer = Committer{}

type Committer struct {
	*tmtypes.ValidatorSet
	Height         uint64
	NextValSetHash []byte
}

func (c Committer) ClientType() exported.ClientType {
	return exported.Tendermint
}

func (c Committer) GetHeight() uint64 {
	return c.Height
}
