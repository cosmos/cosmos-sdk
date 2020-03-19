package module

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

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

type OperationMsg interface {
	Route() string
	Name() string
	Comment() string
	OK() bool
	Msg() json.RawMessage

	LogEvent(eventLogger func(route, op, evResult string))
	MustMarshal() json.RawMessage
}

type FutureOperation interface {
	BlockHeight() int
	BlockTime() time.Time
	Op() Operation
}

// AppParams defines a flat JSON of key/values for all possible configurable
// simulation parameters. It might contain: operation weights, simulation parameters
// and flattened module state parameters (i.e not stored under it's respective module name).
type AppParams interface {
	GetOrGenerate(cdc *codec.Codec, key string, ptr interface{}, r *rand.Rand, ps ParamSimulator)
}

type ParamSimulator func(r *rand.Rand)

type SelectOpFn func(r *rand.Rand) Operation
