package errors

import (
	abci "github.com/tendermint/tendermint/abci/types"
)

// ResponseCheckTx returns an ABCI ResponseCheckTx object with fields filled in
// from the given error and gas values.
func ResponseCheckTx(err error, gw, gu uint64, debug bool) abci.ResponseCheckTx {
	space, code, log := ABCIInfo(err, debug)
	return abci.ResponseCheckTx{
		Codespace: space,
		Code:      code,
		Log:       log,
		GasWanted: int64(gw),
		GasUsed:   int64(gu),
	}
}

// ResponseDeliverTx returns an ABCI ResponseDeliverTx object with fields filled in
// from the given error and gas values.
func ResponseDeliverTx(err error, gw, gu uint64, debug bool) abci.ResponseDeliverTx {
	space, code, log := ABCIInfo(err, debug)
	return abci.ResponseDeliverTx{
		Codespace: space,
		Code:      code,
		Log:       log,
		GasWanted: int64(gw),
		GasUsed:   int64(gu),
	}
}

// QueryResult returns a ResponseQuery from an error. It will try to parse ABCI
// info from the error.
func QueryResult(err error, debug bool) abci.ResponseQuery {
	space, code, log := ABCIInfo(err, debug)
	return abci.ResponseQuery{
		Codespace: space,
		Code:      code,
		Log:       log,
	}
}
