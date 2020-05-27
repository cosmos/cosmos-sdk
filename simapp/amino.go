// +build test_amino

package simapp

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

func AminoJSONTxDecoder(cfg params.EncodingConfig) types.TxDecoder {
	return func(txBytes []byte) (types.Tx, error) {
		var tx authtypes.StdTx
		err := cfg.Marshaler.UnmarshalJSON(txBytes, &tx)
		if err != nil {
			return nil, err
		}
		return tx, nil
	}
}

func NewAnteHandler(ak auth.AccountKeeper, bk bank.Keeper, ibcK ibc.Keeper) types.AnteHandler {
	return ante.NewAnteHandler(
		ak, bk, ibcK, ante.DefaultSigVerificationGasConsumer,
	)
}
