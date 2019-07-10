package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
)

func createFakeTxBuilder() auth.TxBuilder {
	cdc := codec.New()
	return auth.NewTxBuilder(
		utils.GetTxEncoder(cdc),
		123,
		9876,
		0,
		1.2,
		false,
		"test_chain",
		"hello",
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))),
		sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10000, sdk.Precision))},
	)
}

func Test_splitAndCall_NoMessages(t *testing.T) {
	ctx := context.CLIContext{}
	txBldr := createFakeTxBuilder()

	err := splitAndApply(nil, ctx, txBldr, nil, 10)
	assert.NoError(t, err, "")
}

func Test_splitAndCall_Splitting(t *testing.T) {
	ctx := context.CLIContext{}
	txBldr := createFakeTxBuilder()

	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// Add five messages
	msgs := []sdk.Msg{
		sdk.NewTestMsg(addr),
		sdk.NewTestMsg(addr),
		sdk.NewTestMsg(addr),
		sdk.NewTestMsg(addr),
		sdk.NewTestMsg(addr),
	}

	// Keep track of number of calls
	const chunkSize = 2

	callCount := 0
	err := splitAndApply(
		func(ctx context.CLIContext, txBldr auth.TxBuilder, msgs []sdk.Msg) error {
			callCount++

			assert.NotNil(t, ctx)
			assert.NotNil(t, txBldr)
			assert.NotNil(t, msgs)

			if callCount < 3 {
				assert.Equal(t, len(msgs), 2)
			} else {
				assert.Equal(t, len(msgs), 1)
			}

			return nil
		},
		ctx, txBldr, msgs, chunkSize)

	assert.NoError(t, err, "")
	assert.Equal(t, 3, callCount)
}
