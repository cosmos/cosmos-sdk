package v1

import (
	coretransaction "cosmossdk.io/core/transaction"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
)

var _ AppEntrypoint = SimDeliverFn(nil)

type (
	AppEntrypointFn = SimDeliverFn
	SimDeliverFn    func(_txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error)
)

func (m SimDeliverFn) SimDeliver(txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	return m(txEncoder, tx)
}

// testing only
func txConfig() client.TxConfig {
	ir := must(codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          address.NewBech32Codec("cosmos"),
			ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
		},
	}))
	std.RegisterInterfaces(ir)
	ir.RegisterImplementations((*coretransaction.Msg)(nil), &testdata.TestMsg{})
	protoCodec := codec.NewProtoCodec(ir)
	signingCtx := protoCodec.InterfaceRegistry().SigningContext()
	return tx.NewTxConfig(protoCodec, signingCtx.AddressCodec(), signingCtx.ValidatorAddressCodec(), tx.DefaultSignModes)
}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}
