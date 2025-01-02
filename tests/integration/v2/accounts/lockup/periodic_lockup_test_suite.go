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
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestPeriodicLockingAccount() {
	t := s.T()
	currentTime := time.Now()
	ctx := s.ctx
	ctx = integration.SetHeaderInfo(ctx, header.Info{Time: currentTime})

	s.setupStakingParams(ctx, s.stakingKeeper)

	ownerAddrStr, err := s.authKeeper.AddressCodec().BytesToString(accOwner)
	require.NoError(t, err)
	s.fundAccount(s.bankKeeper, ctx, accOwner, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000000))})
	randAcc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	_, accountAddr, err := s.accountsKeeper.Init(ctx, lockupaccount.PERIODIC_LOCKING_ACCOUNT, accOwner, &types.MsgInitPeriodicLockingAccount{
		Owner:     ownerAddrStr,
		StartTime: currentTime,
		LockingPeriods: []types.Period{
			{
				Amount: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(500))),
				Length: time.Minute,
			},
			{
				Amount: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(500))),
				Length: time.Minute,
			},
			{
				Amount: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(500))),
				Length: time.Minute,
			},
		},
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1500))}, nil)
	require.NoError(t, err)

	addr, err := s.authKeeper.AddressCodec().BytesToString(randAcc)
	require.NoError(t, err)

	vals, err := s.stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	val := vals[0]

	t.Run("error - execute message, wrong sender", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    addr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NotNil(t, err)
	})
	// No token being unlocked yet
	t.Run("error - execute send message, insufficient fund", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NotNil(t, err)
	})

	// Update context time
	// After first period 500stake should be unlock
	ctx = integration.SetHeaderInfo(ctx, header.Info{Time: currentTime.Add(time.Minute)})

	// Check if 500 stake is sendable now
	t.Run("ok - execute send message", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(500))},
		}
		err := s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NoError(t, err)

		balance := s.bankKeeper.GetBalance(ctx, randAcc, "stake")
		require.True(t, balance.Amount.Equal(math.NewInt(500)))
	})

	// Update context time
	// After second period 1000stake should be unlock
	ctx = integration.SetHeaderInfo(ctx, header.Info{Time: currentTime.Add(time.Minute * 2)})

	t.Run("ok - execute delegate message", func(t *testing.T) {
		msg := &types.MsgDelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           sdk.NewCoin("stake", math.NewInt(100)),
		}
		err = s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NoError(t, err)

		valbz, err := s.stakingKeeper.ValidatorAddressCodec().StringToBytes(val.OperatorAddress)
		require.NoError(t, err)

		del, err := s.stakingKeeper.Delegations.Get(
			ctx, collections.Join(sdk.AccAddress(accountAddr), sdk.ValAddress(valbz)),
		)
		require.NoError(t, err)
		require.NotNil(t, del)

		// check if tracking is updated accordingly
		lockupAccountInfoResponse := s.queryLockupAccInfo(ctx, s.accountsKeeper, accountAddr)
		delLocking := lockupAccountInfoResponse.DelegatedLocking
		require.True(t, delLocking.AmountOf("stake").Equal(math.NewInt(100)))
	})
	t.Run("ok - execute withdraw reward message", func(t *testing.T) {
		msg := &types.MsgWithdrawReward{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
		}
		err = s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NoError(t, err)
	})
	t.Run("ok - execute undelegate message", func(t *testing.T) {
		vals, err := s.stakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
		msg := &types.MsgUndelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           sdk.NewCoin("stake", math.NewInt(100)),
		}
		err = s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NoError(t, err)
		valbz, err := s.stakingKeeper.ValidatorAddressCodec().StringToBytes(val.OperatorAddress)
		require.NoError(t, err)

		ubd, err := s.stakingKeeper.GetUnbondingDelegation(
			ctx, sdk.AccAddress(accountAddr), sdk.ValAddress(valbz),
		)
		require.NoError(t, err)
		require.Equal(t, len(ubd.Entries), 1)

		// check if an entry is added
		unbondingEntriesResponse := s.queryUnbondingEntries(ctx, s.accountsKeeper, accountAddr, val.OperatorAddress)
		entries := unbondingEntriesResponse.UnbondingEntries
		require.True(t, entries[0].Amount.Amount.Equal(math.NewInt(100)))
		require.True(t, entries[0].ValidatorAddress == val.OperatorAddress)
	})

	// Update context time
	// After third period 1500stake should be unlock
	ctx = integration.SetHeaderInfo(ctx, header.Info{Time: currentTime.Add(time.Minute * 3)})

	// trigger endblock for staking to handle matured unbonding delegation
	_, err = s.stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	t.Run("ok - execute delegate message", func(t *testing.T) {
		msg := &types.MsgDelegate{
			Sender:           ownerAddrStr,
			ValidatorAddress: val.OperatorAddress,
			Amount:           sdk.NewCoin("stake", math.NewInt(100)),
		}
		err = s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NoError(t, err)

		valbz, err := s.stakingKeeper.ValidatorAddressCodec().StringToBytes(val.OperatorAddress)
		require.NoError(t, err)

		del, err := s.stakingKeeper.Delegations.Get(
			ctx, collections.Join(sdk.AccAddress(accountAddr), sdk.ValAddress(valbz)),
		)
		require.NoError(t, err)
		require.NotNil(t, del)

		// check if tracking is updated accordingly
		lockupAccountInfoResponse := s.queryLockupAccInfo(ctx, s.accountsKeeper, accountAddr)
		// check if matured ubd entry cleared
		delLocking := lockupAccountInfoResponse.DelegatedLocking
		require.True(t, delLocking.AmountOf("stake").Equal(math.ZeroInt()))
		delFree := lockupAccountInfoResponse.DelegatedFree
		require.True(t, delFree.AmountOf("stake").Equal(math.NewInt(100)))

		// check if the entry is removed
		unbondingEntriesResponse := s.queryUnbondingEntries(ctx, s.accountsKeeper, accountAddr, val.OperatorAddress)
		entries := unbondingEntriesResponse.UnbondingEntries
		require.Len(t, entries, 0)
	})
}
