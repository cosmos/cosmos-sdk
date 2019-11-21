package tendermint

import (
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ exported.Committer = Committer{}

type Committer struct {
	*tmtypes.ValidatorSet
}

func (c Committer) ClientType() exported.ClientType {
	return exported.Tendermint
}
