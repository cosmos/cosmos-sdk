package errors

import (
	"math"

	abci "github.com/cometbft/cometbft/v2/abci/types"

	errorsmod "cosmossdk.io/errors"
)

// safeInt64FromUint64 converts uint64 to int64 with overflow checking.
// If the value is too large to fit in int64, it returns math.MaxInt64.
func safeInt64FromUint64(val uint64) int64 {
	if val > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(val)
}

// ResponseCheckTxWithEvents returns an ABCI ResponseCheckTx object with fields filled in
// from the given error, gas values and events.
func ResponseCheckTxWithEvents(err error, gw, gu uint64, events []abci.Event, debug bool) *abci.CheckTxResponse {
	space, code, log := errorsmod.ABCIInfo(err, debug)
	return &abci.CheckTxResponse{
		Codespace: space,
		Code:      code,
		Log:       log,
		GasWanted: safeInt64FromUint64(gw),
		GasUsed:   safeInt64FromUint64(gu),
		Events:    events,
	}
}

// ResponseExecTxResultWithEvents returns an ABCI ExecTxResult object with fields
// filled in from the given error, gas values and events.
func ResponseExecTxResultWithEvents(err error, gw, gu uint64, events []abci.Event, debug bool) *abci.ExecTxResult {
	space, code, log := errorsmod.ABCIInfo(err, debug)
	return &abci.ExecTxResult{
		Codespace: space,
		Code:      code,
		Log:       log,
		GasWanted: safeInt64FromUint64(gw),
		GasUsed:   safeInt64FromUint64(gu),
		Events:    events,
	}
}

// QueryResult returns a ResponseQuery from an error. It will try to parse ABCI
// info from the error.
func QueryResult(err error, debug bool) *abci.QueryResponse {
	space, code, log := errorsmod.ABCIInfo(err, debug)
	return &abci.QueryResponse{
		Codespace: space,
		Code:      code,
		Log:       log,
	}
}
