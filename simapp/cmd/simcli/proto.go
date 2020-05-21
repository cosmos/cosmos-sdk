// +build !test_amino

package main

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func MakeTxCLIContext() context.CLIContext {
	cliCtx := context.CLIContext{}
	protoCdc := codec.NewProtoCodec(interfaceRegistry)
	return cliCtx.
		WithJSONMarshaler(protoCdc).
		WithTxGenerator(signing.TxGenerator{Marshaler: protoCdc}).
		WithAccountRetriever(types.NewAccountRetriever(appCodec)).
		WithCodec(cdc)
}
