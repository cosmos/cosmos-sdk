// +build !test_amino

package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func MakeTxCLIContext() client.Context {
	cliCtx := client.Context{}
	protoCdc := codec.NewProtoCodec(encodingConfig.InterfaceRegistry)
	return cliCtx.
		WithJSONMarshaler(protoCdc).
		WithTxGenerator(encodingConfig.TxGenerator).
		WithAccountRetriever(types.NewAccountRetriever(encodingConfig.Marshaler)).
		WithCodec(encodingConfig.Amino)
}
