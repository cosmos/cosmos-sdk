package module

import (
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WeightedProposalContent interface {
	AppParamsKey() string                   // key used to retrieve the value of the weight from the simulation application params
	DefaultWeight() int                     // default weight
	ContentSimulatorFn() ContentSimulatorFn // content simulator function
}

type ContentSimulatorFn func(r *rand.Rand, ctx sdk.Context, accs []Account) Content

type Content interface {
	GetTitle() string
	GetDescription() string
	ProposalRoute() string
	ProposalType() string
	ValidateBasic() error
	String() string
}

type Account interface {
	PrivKey() crypto.PrivKey
	PubKey() crypto.PubKey
	Address() sdk.AccAddress
}
