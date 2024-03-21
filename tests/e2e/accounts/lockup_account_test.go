//go:build app_v1

package accounts

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	lockupaccount "cosmossdk.io/x/accounts/lockup"
	"cosmossdk.io/x/accounts/lockup/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	accOwner = secp256k1.GenPrivKey().PubKey().Address()
)

func TestContinuousLockingAccount(t *testing.T) {
	app := setupApp(t)
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())
	currentTime := time.Now()
	ownerAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(accOwner)
	require.NoError(t, err)
	fundAccount(t, app, ctx, ownerAddrStr, "1000000stake")

	_, accountAddr, err := app.AccountsKeeper.Init(ctx, lockupaccount.CONTINUOUS_LOCKING_ACCOUNT, accOwner, &types.MsgInitLockupAccount{
		Owner:     accOwner.string,
		StartTime: currentTime,
		// end time in 1 minutes
		EndTime: currentTime.Add(time.Minute * 1),
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	excuteTx(t, ctx, msg, ak, accountAddr, accOwner)
	t.Run("ok - execute send message", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: bechify(t, app, []byte("random-addr")),
			Amount:    coins(t, "100stake"),
		}
		excuteTx(t, ctx, msg, app, accountAddr, accOwner)
	})
	t.Run("ok - execute delegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
		msg := &types.MsgDelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           coins(t, "100stake"),
		}
		excuteTx(t, ctx, msg, app, accountAddr, accOwner)
	})
	t.Run("ok - execute undelegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
		msg := &types.MsgUndelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           coins(t, "100stake"),
		}
		excuteTx(t, ctx, msg, app, accountAddr, accOwner)
	})
}

func TestDelayedLockingAccount(t *testing.T) {
	app := setupApp(t)
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())
	fundAccount(t, app, ctx, accCreator, "1000000stake")

	_, accountAddr, err := app.AccountsKeeper.Init(ctx, lockupaccount.DELAYED_LOCKING_ACCOUNT, accCreator, &types.MsgInitLockupAccount{
		Owner: accOwner,
		// end time in 1 minutes
		EndTime: time.Now().Add(time.Minute * 1),
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	t.Run("ok - execute send message", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: bechify(t, app, []byte("random-addr")),
			Amount:    coins(t, "100stake"),
		}
		excuteTx(t, ctx, msg, app, accountAddr, accOwner)
	})
	t.Run("ok - execute delegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
		msg := &types.MsgDelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           coins(t, "100stake"),
		}
		excuteTx(t, ctx, msg, app, accountAddr, accOwner)
	})
	t.Run("ok - execute undelegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
		msg := &types.MsgUndelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           coins(t, "100stake"),
		}
		excuteTx(t, ctx, msg, app, accountAddr, accOwner)
	})
}

func TestPeriodicLockingAccount(t *testing.T) {
	app := setupApp(t)
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())
	currentTime := time.Now()
	fundAccount(t, app, ctx, accCreator, "1000000stake")

	_, accountAddr, err := app.AccountsKeeper.Init(ctx, lockupaccount.PERIODIC_LOCKING_ACCOUNT, accCreator, &types.MsgInitLockupAccount{
		Owner:     accOwner,
		StartTime: currentTime,
		// end time in 1 minutes
		EndTime: time.Now().Add(time.Minute * 1),
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	t.Run("ok - execute send message", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: bechify(t, app, []byte("random-addr")),
			Amount:    coins(t, "100stake"),
		}
		excuteTx(t, ctx, msg, app, accountAddr, accOwner)
	})
	t.Run("ok - execute delegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
		msg := &types.MsgDelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           coins(t, "100stake"),
		}
		excuteTx(t, ctx, msg, app, accountAddr, accOwner)
	})
	t.Run("ok - execute undelegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
		msg := &types.MsgUndelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           coins(t, "100stake"),
		}
		excuteTx(t, ctx, msg, app, accountAddr, accOwner)
	})
}

func executeTx(t *testing.T, ctx sdk.Context, msg sdk.Msg, app *simapp.SimApp, sender, accAddr []byte) {
	_, err := app.AccountsKeeper.Execute(ctx, accAddr, sender, msg, nil)
	require.NoError(t, err)
}
