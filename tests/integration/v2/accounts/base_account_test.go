package accounts

import (
	"context"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	gogoany "github.com/cosmos/gogoproto/types/any"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/accounts"
	baseaccountv1 "cosmossdk.io/x/accounts/defaults/base/v1"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

var (
	privKey    = secp256k1.GenPrivKey()
	accCreator = []byte("creator")
)

func TestBaseAccount(t *testing.T) {
	f := initFixture(t, nil)
	app := f.app
	ctx := f.ctx

	_, baseAccountAddr, err := f.accountsKeeper.Init(ctx, "base", accCreator, &baseaccountv1.MsgInit{
		PubKey: toAnyPb(t, privKey.PubKey()),
	}, nil, nil)
	require.NoError(t, err)

	// fund base account! this will also cause an auth base account to be created
	// by the bank module.
	fundAccount(t, f.bankKeeper, ctx, baseAccountAddr, "1000000stake")

	// now we make the account send a tx, public key not present.
	// so we know it will default to x/accounts calling.
	msg := &banktypes.MsgSend{
		FromAddress: bechify(t, f.authKeeper, baseAccountAddr),
		ToAddress:   bechify(t, f.authKeeper, []byte("random-addr")),
		Amount:      coins(t, "100stake"),
	}
	sendTx(t, ctx, app, f.accountsKeeper, baseAccountAddr, msg)
}

func sendTx(t *testing.T, ctx context.Context, app *integration.App, ak accounts.Keeper, sender []byte, msg sdk.Msg) {
	t.Helper()
	accNum, err := ak.AccountByNumber.Get(ctx, sender)
	require.NoError(t, err)

	accSeq, err := ak.Query(ctx, sender, &baseaccountv1.QuerySequence{})
	require.NoError(t, err)

	app.SignCheckDeliver(
		t, ctx, []sdk.Msg{msg}, "", []uint64{accNum}, []uint64{accSeq.(*baseaccountv1.QuerySequenceResponse).Sequence},
		[]cryptotypes.PrivKey{privKey},
		"",
	)
}

func bechify(t *testing.T, ak authkeeper.AccountKeeper, addr []byte) string {
	t.Helper()
	bech32, err := ak.AddressCodec().BytesToString(addr)
	require.NoError(t, err)
	return bech32
}

func fundAccount(t *testing.T, bk bankkeeper.Keeper, ctx context.Context, addr sdk.AccAddress, amt string) {
	t.Helper()
	require.NoError(t, testutil.FundAccount(ctx, bk, addr, coins(t, amt)))
}

func toAnyPb(t *testing.T, pm gogoproto.Message) *codectypes.Any {
	t.Helper()
	if gogoproto.MessageName(pm) == gogoproto.MessageName(&gogoany.Any{}) {
		t.Fatal("no")
	}
	pb, err := codectypes.NewAnyWithValue(pm)
	require.NoError(t, err)
	return pb
}

func coins(t *testing.T, s string) sdk.Coins {
	t.Helper()
	coins, err := sdk.ParseCoinsNormalized(s)
	require.NoError(t, err)
	return coins
}
