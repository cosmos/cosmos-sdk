package app

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// This will fail half the time with the second output being 173
// This is due to secp256k1 signatures not being constant size.
// This will be resolved when updating to tendermint v0.24.0
// nolint: vet
func ExampleTxSendSize() {
	cdc := app.MakeCodec()
	priv1 := secp256k1.GenPrivKeySecp256k1([]byte{0})
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	priv2 := secp256k1.GenPrivKeySecp256k1([]byte{1})
	addr2 := sdk.AccAddress(priv2.PubKey().Address())
	coins := []sdk.Coin{sdk.NewCoin("denom", sdk.NewInt(10))}
	msg1 := bank.MsgSend{
		Inputs:  []bank.Input{bank.NewInput(addr1, coins)},
		Outputs: []bank.Output{bank.NewOutput(addr2, coins)},
	}
	sig, _ := priv1.Sign(msg1.GetSignBytes())
	sigs := []auth.StdSignature{{nil, sig, 0, 0}}
	tx := auth.NewStdTx([]sdk.Msg{msg1}, auth.NewStdFee(0, coins...), sigs, "")
	fmt.Println(len(cdc.MustMarshalBinaryBare([]sdk.Msg{msg1})))
	fmt.Println(len(cdc.MustMarshalBinaryBare(tx)))
	// output: 80
	// 167
}
