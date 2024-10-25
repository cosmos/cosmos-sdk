package lockup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	lockupaccount "cosmossdk.io/x/accounts/defaults/lockup"
	types "cosmossdk.io/x/accounts/defaults/lockup/v1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *E2ETestSuite) TestPermanentLockingAccount() {
	t := s.T()
	app := setupApp(t)
	currentTime := time.Now()
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger()).WithHeaderInfo(header.Info{
		Time: currentTime,
	})
	ownerAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(accOwner)
	require.NoError(t, err)
	s.fundAccount(app, ctx, accOwner, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000000))})
	randAcc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	_, accountAddr, err := app.AccountsKeeper.Init(ctx, lockupaccount.PERMANENT_LOCKING_ACCOUNT, accOwner, &types.MsgInitLockupAccount{
		Owner: ownerAddrStr,
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	addr, err := app.AuthKeeper.AddressCodec().BytesToString(randAcc)
	require.NoError(t, err)

	vals, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	val := vals[0]

	t.Run("error - execute message, wrong sender", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    addr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NotNil(t, err)
	})
	t.Run("error - execute send message, insufficient fund", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NotNil(t, err)
	})
	t.Run("ok - execute delegate message", func(t *testing.T) {
		msg := &types.MsgDelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           sdk.NewCoin("stake", math.NewInt(100)),
		}
		err = s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NoError(t, err)

		valbz, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.OperatorAddress)
		require.NoError(t, err)

		del, err := app.StakingKeeper.Delegations.Get(
			ctx, collections.Join(sdk.AccAddress(accountAddr), sdk.ValAddress(valbz)),
		)
		require.NoError(t, err)
		require.NotNil(t, del)

		// check if tracking is updated accordingly
		lockupAccountInfoResponse := s.queryLockupAccInfo(ctx, app, accountAddr)
		delLocking := lockupAccountInfoResponse.DelegatedLocking
		require.True(t, delLocking.AmountOf("stake").Equal(math.NewInt(100)))
	})
	t.Run("ok - execute withdraw reward message", func(t *testing.T) {
		msg := &types.MsgWithdrawReward{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
		}
		err = s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NoError(t, err)
	})
	t.Run("ok - execute undelegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
		msg := &types.MsgUndelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           sdk.NewCoin("stake", math.NewInt(100)),
		}
		err = s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NoError(t, err)
		valbz, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.OperatorAddress)
		require.NoError(t, err)

		ubd, err := app.StakingKeeper.GetUnbondingDelegation(
			ctx, sdk.AccAddress(accountAddr), sdk.ValAddress(valbz),
		)
		require.NoError(t, err)
		require.Equal(t, len(ubd.Entries), 1)

		// check if tracking is updated accordingly
		lockupAccountInfoResponse := s.queryLockupAccInfo(ctx, app, accountAddr)
		delLocking := lockupAccountInfoResponse.DelegatedLocking
		require.True(t, delLocking.AmountOf("stake").Equal(math.ZeroInt()))
	})

	s.fundAccount(app, ctx, accountAddr, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})

	t.Run("ok - execute send message", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NoError(t, err)

		balance := app.BankKeeper.GetBalance(ctx, randAcc, "stake")
		require.True(t, balance.Amount.Equal(math.NewInt(100)))
	})
}
