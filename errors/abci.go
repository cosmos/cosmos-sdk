package errors

import (
	abci "github.com/tendermint/abci/types"
)

type causer interface {
	Cause() error
}

func getABCIError(err error) (ABCIError, bool) {
	if err, ok := err.(ABCIError); ok {
		return err, true
	}
	if causer, ok := err.(causer); ok {
		err := causer.Cause()
		if err, ok := err.(ABCIError); ok {
			return err, true
		}
	}
	return nil, false
}

func ResponseDeliverTxFromErr(err error) *abci.ResponseDeliverTx {
	var code = CodeInternalError
	var log = codeToDefaultLog(code)

	abciErr, ok := getABCIError(err)
	if ok {
		code = abciErr.ABCICode()
		log = abciErr.ABCILog()
	}

	return &abci.ResponseDeliverTx{
		Code: code,
		Data: nil,
		Log:  log,
		Tags: nil,
	}
}

func ResponseCheckTxFromErr(err error) *abci.ResponseCheckTx {
	var code = CodeInternalError
	var log = codeToDefaultLog(code)

	abciErr, ok := getABCIError(err)
	if ok {
		code = abciErr.ABCICode()
		log = abciErr.ABCILog()
	}

	return &abci.ResponseCheckTx{
		Code: code,
		Data: nil,
		Log:  log,
		Gas:  0, // TODO
		Fee:  0, // TODO
	}
}
