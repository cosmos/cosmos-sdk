//go:build app_v1

package accounts

import (
	"context"
	"testing"

	"cosmossdk.io/simapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

var (
	privKey     = secp256k1.GenPrivKey()
	accCreator  = []byte("creator")
	bundlerAddr = secp256k1.GenPrivKey().PubKey().Address()
	aliceAddr   = secp256k1.GenPrivKey().PubKey().Address()
)

/*
func TestAccountAbstraction(t *testing.T) {
	app := setupApp(t)
	ak := app.AccountsKeeper
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())

	_, aaAddr, err := ak.Init(ctx, "aa_minimal", accCreator, &rotationv1.MsgInit{
		PubKeyBytes: privKey.PubKey().Bytes(),
	}, nil)
	require.NoError(t, err)

	_, aaFullAddr, err := ak.Init(ctx, "aa_full", accCreator, &rotationv1.MsgInit{
		PubKeyBytes: privKey.PubKey().Bytes(),
	}, nil)
	require.NoError(t, err)

	aaAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(aaAddr)
	require.NoError(t, err)

	aaFullAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(aaFullAddr)
	require.NoError(t, err)

	// let's give aa some coins.
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, aaAddr, sdk.NewCoins(sdk.NewInt64Coin("stake", 100000000000))))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, aaFullAddr, sdk.NewCoins(sdk.NewInt64Coin("stake", 100000000000))))

	bundlerAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(bundlerAddr)
	require.NoError(t, err)

	aliceAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(aliceAddr)
	require.NoError(t, err)

	t.Run("ok - pay bundler not implemented", func(t *testing.T) {})
	t.Run("pay bundle impersonation", func(t *testing.T) {})
	t.Run("auth failure", func(t *testing.T) {})
	t.Run("pay bundle failure", func(t *testing.T) {})
	t.Run("exec message failure", func(t *testing.T) {})

	t.Run("implements bundler payment - fail ", func(t *testing.T) {})

	t.Run("implements execution - fail", func(t *testing.T) {})

	t.Run("implements bundler payment and execution - success", func(t *testing.T) {})

	t.Run("Simulate - OK", func(t *testing.T) {})

	t.Run("Simulate - Fail empty user operation", func(t *testing.T) {})
}
*/

func intoAny(t *testing.T, msgs ...gogoproto.Message) (anys []*codectypes.Any) {
	t.Helper()
	for _, msg := range msgs {
		any, err := codectypes.NewAnyWithValue(msg)
		require.NoError(t, err)
		anys = append(anys, any)
	}
	return
}

func coins(t *testing.T, s string) sdk.Coins {
	t.Helper()
	coins, err := sdk.ParseCoinsNormalized(s)
	require.NoError(t, err)
	return coins
}

func balanceIs(t *testing.T, ctx context.Context, app *simapp.SimApp, addr sdk.AccAddress, s string) {
	t.Helper()
	balance := app.BankKeeper.GetAllBalances(ctx, addr)
	require.Equal(t, s, balance.String())
}

var mockSignature = &codectypes.Any{TypeUrl: "signature", Value: []byte("signature")}

func setupApp(t *testing.T) *simapp.SimApp {
	t.Helper()
	app := simapp.Setup(t, false)
	return app
}
