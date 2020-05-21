// +build !test_amino

package main

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func MakeTxCLIContext() context.CLIContext {
	cliCtx := context.CLIContext{}
	return cliCtx.
		WithJSONMarshaler(appCodec).
		WithTxGenerator(types.StdTxGenerator{Cdc: cdc}).
		WithAccountRetriever(types.NewAccountRetriever(appCodec)).
		WithCodec(cdc)
}
