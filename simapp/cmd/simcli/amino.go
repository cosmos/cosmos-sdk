// +build test_amino

package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func MakeTxCLIContext() client.Context {
	cliCtx := client.Context{}
	aminoCdc := codec.NewAminoCodec(encodingConfig.Amino)
	return cliCtx.
		WithJSONMarshaler(aminoCdc).
		WithTxGenerator(encodingConfig.TxGenerator).
		WithAccountRetriever(types.NewAccountRetriever(encodingConfig.Marshaler)).
		WithCodec(encodingConfig.Amino)
}
