//go:build app_v1

package accounts

import (
	"math/rand"
	"testing"

	"cosmossdk.io/simapp"
	baseaccountv1 "cosmossdk.io/x/accounts/defaults/base/v1"
	banktypes "cosmossdk.io/x/bank/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestBaseAccount(t *testing.T) {
	app := setupApp(t)
	ak := app.AccountsKeeper
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())

	_, baseAccountAddr, err := ak.Init(ctx, "base", accCreator, &baseaccountv1.MsgInit{
		PubKey: privKey.PubKey().Bytes(),
	})
	require.NoError(t, err)

	// fund base account! this will also cause an auth base account to be created
	// by the bank module.
	fundAccount(t, app, ctx, baseAccountAddr, "1000000stake")

	// now we make the account send a tx, public key not present.
	// so we know it will default to x/accounts calling.
	msg := &banktypes.MsgSend{
		FromAddress: bechify(t, app, baseAccountAddr),
		ToAddress:   bechify(t, app, []byte("random-addr")),
		Amount:      coins(t, "100stake"),
	}
	tx := sign(t, ctx, app, baseAccountAddr, privKey, msg)
	res, _, err := app.SimDeliver(app.TxEncode, tx)
	require.NoError(t, err)
	t.Log(res)
}

func sign(t *testing.T, ctx sdk.Context, app *simapp.SimApp, from sdk.AccAddress, privKey cryptotypes.PrivKey, msg sdk.Msg) sdk.Tx {
	r := rand.New(rand.NewSource(0))

	accNum, err := app.AccountsKeeper.AccountByNumber.Get(ctx, from)
	require.NoError(t, err)
	accSeq, err := app.AccountsKeeper.Query(ctx, from, &baseaccountv1.QuerySequence{})
	require.NoError(t, err)

	tx, err := sims.GenSignedMockTx(
		r,
		app.TxConfig(),
		[]sdk.Msg{msg},
		coins(t, "100stake"),
		1_000_000,
		app.ChainID(),
		[]uint64{accNum},
		[]uint64{accSeq.(*baseaccountv1.QuerySequenceResponse).Sequence},
		privKey,
	)

	require.NoError(t, err)
	return tx
}

func bechify(t *testing.T, app *simapp.SimApp, addr []byte) string {
	bech32, err := app.AuthKeeper.AddressCodec().BytesToString(addr)
	require.NoError(t, err)
	return bech32
}
