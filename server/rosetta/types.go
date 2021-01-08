package rosetta

import (
	"github.com/coinbase/rosetta-sdk-go/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// statuses
const (
	StatusSuccess  = "Success"
	StatusReverted = "Reverted"
	StageSynced    = "synced"
	StageSyncing   = "syncing"
)

// misc
const (
	Log = "log"
)

// operations
const (
	OperationFee = "fee"
)

// options
const (
	OptionAccountNumber = "account_number"
	OptionAddress       = "address"
	OptionChainID       = "chain_id"
	OptionSequence      = "sequence"
	OptionMemo          = "memo"
	OptionGas           = "gas"
)

type Msg interface {
	sdk.Msg
	ToOperations(withStatus, hasError bool) []*types.Operation
	FromOperations(ops []*types.Operation) (sdk.Msg, error)
}
