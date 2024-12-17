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

func (s *IntegrationTestSuite) TestDelayedLockingAccount() {
	t := s.T()
	currentTime := time.Now()
	ctx := s.ctx
	ctx = integration.SetHeaderInfo(ctx, header.Info{Time: currentTime})

	ownerAddrStr, err := s.authKeeper.AddressCodec().BytesToString(accOwner)
	require.NoError(t, err)
	s.fundAccount(s.bankKeeper, ctx, accOwner, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000000))})
	randAcc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	withdrawAcc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	_, accountAddr, err := s.accountsKeeper.Init(ctx, lockupaccount.DELAYED_LOCKING_ACCOUNT, accOwner, &types.MsgInitLockupAccount{
		Owner: ownerAddrStr,
		// end time in 1 minutes
		EndTime: currentTime.Add(time.Minute),
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))}, nil)
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
	t.Run("error - execute send message, insufficient fund", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NotNil(t, err)
	})
	t.Run("error - execute withdraw message, no withdrawable token", func(t *testing.T) {
		ownerAddr, err := s.authKeeper.AddressCodec().BytesToString(accOwner)
		require.NoError(t, err)
		withdrawAddr, err := s.authKeeper.AddressCodec().BytesToString(withdrawAcc)
		require.NoError(t, err)
		msg := &types.MsgWithdraw{
			Withdrawer: ownerAddr,
			ToAddress:  withdrawAddr,
			Denoms:     []string{"stake"},
		}
		err = s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NotNil(t, err)
	})
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

		// check if tracking is updated accordingly
		lockupAccountInfoResponse := s.queryLockupAccInfo(ctx, s.accountsKeeper, accountAddr)
		delLocking := lockupAccountInfoResponse.DelegatedLocking
		require.True(t, delLocking.AmountOf("stake").Equal(math.ZeroInt()))
	})

	// Update context time
	// After endtime fund should be unlock
	ctx = integration.SetHeaderInfo(ctx, header.Info{Time: currentTime.Add(time.Second * 61)})

	// Check if token is sendable after unlock
	t.Run("ok - execute send message", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NoError(t, err)

		balance := s.bankKeeper.GetBalance(ctx, randAcc, "stake")
		require.True(t, balance.Amount.Equal(math.NewInt(100)))
	})
	// Test to withdraw all the remain funds to an account of choice
	t.Run("ok - execute withdraw message", func(t *testing.T) {
		ownerAddr, err := s.authKeeper.AddressCodec().BytesToString(accOwner)
		require.NoError(t, err)
		withdrawAddr, err := s.authKeeper.AddressCodec().BytesToString(withdrawAcc)
		require.NoError(t, err)
		msg := &types.MsgWithdraw{
			Withdrawer: ownerAddr,
			ToAddress:  withdrawAddr,
			Denoms:     []string{"stake"},
		}
		err = s.executeTx(ctx, msg, s.accountsKeeper, accountAddr, accOwner)
		require.NoError(t, err)

		// withdrawable amount should be
		// 1000stake - 100stake( above sent amt ) - 100stake(above delegate amt) = 800stake
		balance := s.bankKeeper.GetBalance(ctx, withdrawAcc, "stake")
		require.True(t, balance.Amount.Equal(math.NewInt(800)))
	})
}
