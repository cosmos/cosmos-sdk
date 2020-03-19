package module

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
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

type ParamChange interface {
	Subspace() string
	Key() string
	SimValue() SimValFn
}

type SimValFn func(r *rand.Rand) string

type WeightedOperation interface {
	Weight() int
	Op() Operation
}

type Operation func(r *rand.Rand, app *baseapp.BaseApp,
	ctx sdk.Context, accounts []Account, chainID string) (
	OperationMsg OperationMsg, futureOps []FutureOperation, err error)

type OperationMsg struct {
	Route   string          `json:"route" yaml:"route"`     // msg route (i.e module name)
	Name    string          `json:"name" yaml:"name"`       // operation name (msg Type or "no-operation")
	Comment string          `json:"comment" yaml:"comment"` // additional comment
	OK      bool            `json:"ok" yaml:"ok"`           // success
	Msg     json.RawMessage `json:"msg" yaml:"msg"`         // JSON encoded msg
}

type FutureOperation struct {
	BlockHeight int
	BlockTime   time.Time
	Op          Operation
}
