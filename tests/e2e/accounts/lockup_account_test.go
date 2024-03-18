package accounts

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	lockupaccount "cosmossdk.io/x/accounts/lockup"
	"cosmossdk.io/x/accounts/lockup/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	bundlerAddr = secp256k1.GenPrivKey().PubKey().Address()
	accOwner    = secp256k1.GenPrivKey().PubKey().Address()
)

func TestContinuousLockingAccount(t *testing.T) {
	app := setupApp(t)
	ak := app.AccountsKeeper
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())
	currentTime := time.Now()
	ownerAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(accOwner)
	require.NoError(t, err)
	fundAccount(t, app, ctx, ownerAddrStr, "1000000stake")

	_, accountAddr, err := ak.Init(ctx, lockupaccount.CONTINUOUS_LOCKING_ACCOUNT, accOwner, &types.MsgInitLockupAccount{
		Owner:     accOwner.string,
		StartTime: currentTime,
		// end time in 1 minutes
		EndTime: time.Now().Add(time.Minute * 1),
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	// now we make the account send a tx, public key not present.
	// so we know it will default to x/accounts calling.
	msg := &types.MsgSend{
		ToAddress: bechify(t, app, []byte("random-addr")),
		Amount:    coins(t, "100stake"),
	}
	sendTx(t, ctx, app, accountAddr, msg)
}

func TestDelayedLockingAccount(t *testing.T) {
	app := setupApp(t)
	ak := app.AccountsKeeper
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())
	currentTime := time.Now()
	fundAccount(t, app, ctx, accCreator, "1000000stake")

	_, accountAddr, err := ak.Init(ctx, lockupaccount.CONTINUOUS_LOCKING_ACCOUNT, accCreator, &types.MsgInitLockupAccount{
		Owner:     accOwner,
		StartTime: currentTime,
		// end time in 1 minutes
		EndTime: time.Now().Add(time.Minute * 1),
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	// now we make the account send a tx, public key not present.
	// so we know it will default to x/accounts calling.
	msg := &types.MsgSend{
		ToAddress: bechify(t, app, []byte("random-addr")),
		Amount:    coins(t, "100stake"),
	}
	sendTx(t, ctx, app, accountAddr, msg)
}

func TestPeriodicLockingAccount(t *testing.T) {
	app := setupApp(t)
	ak := app.AccountsKeeper
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())
	currentTime := time.Now()
	fundAccount(t, app, ctx, accCreator, "1000000stake")

	_, accountAddr, err := ak.Init(ctx, lockupaccount.CONTINUOUS_LOCKING_ACCOUNT, accCreator, &types.MsgInitLockupAccount{
		Owner:     accOwner,
		StartTime: currentTime,
		// end time in 1 minutes
		EndTime: time.Now().Add(time.Minute * 1),
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	// now we make the account send a tx, public key not present.
	// so we know it will default to x/accounts calling.
	msg := &types.MsgSend{
		ToAddress: bechify(t, app, []byte("random-addr")),
		Amount:    coins(t, "100stake"),
	}
	sendTx(t, ctx, app, accountAddr, msg)
}
