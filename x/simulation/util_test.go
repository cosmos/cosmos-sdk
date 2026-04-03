package simulation

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type nilAccountKeeper struct{}

func (nilAccountKeeper) GetAccount(context.Context, sdk.AccAddress) sdk.AccountI { return nil }

func TestGenAndDeliverTx_AccountNotFoundReturnsNoOp(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	txCtx := OperationInput{
		R:          rand.New(rand.NewSource(1)),
		Msg:        &banktypes.MsgSend{},
		ModuleName: banktypes.ModuleName,
		Context:    sdk.Context{},
		SimAccount: simtypes.Account{
			PrivKey: privKey,
			Address: sdk.AccAddress(privKey.PubKey().Address()),
		},
		AccountKeeper: nilAccountKeeper{},
	}

	opMsg, futureOps, err := GenAndDeliverTx(txCtx, nil)
	require.NoError(t, err)
	require.Empty(t, futureOps)
	require.False(t, opMsg.OK)
	require.Equal(t, "account not found", opMsg.Comment)
}
