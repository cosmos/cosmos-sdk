package accounts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	lockupaccount "cosmossdk.io/x/accounts/defaults/lockup"
	"cosmossdk.io/x/accounts/defaults/lockup/types"
	"cosmossdk.io/x/bank/testutil"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ownerAddr = secp256k1.GenPrivKey().PubKey().Address()
	accOwner  = sdk.AccAddress(ownerAddr)
)

type E2ETestSuite struct {
	suite.Suite

	app *simapp.SimApp
}

func NewE2ETestSuite() *E2ETestSuite {
	return &E2ETestSuite{}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")
	s.app = setupApp(s.T())
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
}

func setupApp(t *testing.T) *simapp.SimApp {
	t.Helper()
	app := simapp.Setup(t, false)
	return app
}

func (s *E2ETestSuite) executeTx(ctx sdk.Context, msg sdk.Msg, app *simapp.SimApp, accAddr, sender []byte) error {
	_, err := app.AccountsKeeper.Execute(ctx, accAddr, sender, msg, nil)
	return err
}

func (s *E2ETestSuite) fundAccount(app *simapp.SimApp, ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) {
	require.NoError(s.T(), testutil.FundAccount(ctx, app.BankKeeper, addr, amt))
}

func (s *E2ETestSuite) TestContinuousLockingAccount() {
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
	withdrawAcc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	_, accountAddr, err := app.AccountsKeeper.Init(ctx, lockupaccount.CONTINUOUS_LOCKING_ACCOUNT, accOwner, &types.MsgInitLockupAccount{
		Owner:     ownerAddrStr,
		StartTime: currentTime,
		// end time in 1 minutes
		EndTime: currentTime.Add(time.Minute),
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	addr, err := app.AuthKeeper.AddressCodec().BytesToString(randAcc)
	require.NoError(t, err)

	t.Run("error - execute send message, insufficient fund", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NotNil(t, err)
	})
	t.Run("error - execute withdraw message, no withdrawable token", func(t *testing.T) {
		ownerAddr, err := app.AuthKeeper.AddressCodec().BytesToString(accOwner)
		require.NoError(t, err)
		withdrawAddr, err := app.AuthKeeper.AddressCodec().BytesToString(withdrawAcc)
		require.NoError(t, err)
		msg := &types.MsgWithdraw{
			Withdrawer: ownerAddr,
			ToAddress:  withdrawAddr,
			Denoms:     []string{"stake"},
		}
		err = s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NotNil(t, err)
	})

	// Update context time
	// 12 sec = 1/5 of a minute so 200stake should be released
	ctx = ctx.WithHeaderInfo(header.Info{
		Time: currentTime.Add(time.Second * 12),
	})

	// Check if token is sendable
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
	t.Run("ok - execute withdraw message", func(t *testing.T) {
		ownerAddr, err := app.AuthKeeper.AddressCodec().BytesToString(accOwner)
		require.NoError(t, err)
		withdrawAddr, err := app.AuthKeeper.AddressCodec().BytesToString(withdrawAcc)
		require.NoError(t, err)
		msg := &types.MsgWithdraw{
			Withdrawer: ownerAddr,
			ToAddress:  withdrawAddr,
			Denoms:     []string{"stake"},
		}
		err = s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NoError(t, err)

		// withdrawable amount should be 200 - 100 = 100stake
		balance := app.BankKeeper.GetBalance(ctx, withdrawAcc, "stake")
		require.True(t, balance.Amount.Equal(math.NewInt(100)))
	})
	t.Run("ok - execute delegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
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
	})
}

func (s *E2ETestSuite) TestDelayedLockingAccount() {
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
	withdrawAcc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	_, accountAddr, err := app.AccountsKeeper.Init(ctx, lockupaccount.DELAYED_LOCKING_ACCOUNT, accOwner, &types.MsgInitLockupAccount{
		Owner: ownerAddrStr,
		// end time in 1 minutes
		EndTime: currentTime.Add(time.Minute),
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	addr, err := app.AuthKeeper.AddressCodec().BytesToString(randAcc)
	require.NoError(t, err)

	t.Run("error - execute send message, insufficient fund", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NotNil(t, err)
	})
	t.Run("error - execute withdraw message, no withdrawable token", func(t *testing.T) {
		ownerAddr, err := app.AuthKeeper.AddressCodec().BytesToString(accOwner)
		require.NoError(t, err)
		withdrawAddr, err := app.AuthKeeper.AddressCodec().BytesToString(withdrawAcc)
		require.NoError(t, err)
		msg := &types.MsgWithdraw{
			Withdrawer: ownerAddr,
			ToAddress:  withdrawAddr,
			Denoms:     []string{"stake"},
		}
		err = s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NotNil(t, err)
	})
	t.Run("ok - execute delegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
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
	})

	// Update context time
	// After endtime fund should be unlock
	ctx = ctx.WithHeaderInfo(header.Info{
		Time: currentTime.Add(time.Second * 61),
	})

	// Check if token is sendable after unlock
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
	// Test to withdraw all the remain funds to an account of choice
	t.Run("ok - execute withdraw message", func(t *testing.T) {
		ownerAddr, err := app.AuthKeeper.AddressCodec().BytesToString(accOwner)
		require.NoError(t, err)
		withdrawAddr, err := app.AuthKeeper.AddressCodec().BytesToString(withdrawAcc)
		require.NoError(t, err)
		msg := &types.MsgWithdraw{
			Withdrawer: ownerAddr,
			ToAddress:  withdrawAddr,
			Denoms:     []string{"stake"},
		}
		err = s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NoError(t, err)

		// withdrawable amount should be
		// 1000stake - 100stake( above sent amt ) - 100stake(above delegate amt) = 800stake
		balance := app.BankKeeper.GetBalance(ctx, withdrawAcc, "stake")
		require.True(t, balance.Amount.Equal(math.NewInt(800)))
	})
}

func (s *E2ETestSuite) TestPeriodicLockingAccount() {
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
	withdrawAcc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	_, accountAddr, err := app.AccountsKeeper.Init(ctx, lockupaccount.PERIODIC_LOCKING_ACCOUNT, accOwner, &types.MsgInitPeriodicLockingAccount{
		Owner:     ownerAddrStr,
		StartTime: currentTime,
		LockingPeriods: []types.Period{
			{
				Amount: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(500))),
				Length: time.Minute,
			}, {
				Amount: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(500))),
				Length: time.Minute,
			},
		},
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	addr, err := app.AuthKeeper.AddressCodec().BytesToString(randAcc)
	require.NoError(t, err)

	// No token being unlocked yet
	t.Run("error - execute send message, insufficient fund", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))},
		}
		err := s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NotNil(t, err)
	})
	t.Run("error - execute withdraw message, no withdrawable token", func(t *testing.T) {
		ownerAddr, err := app.AuthKeeper.AddressCodec().BytesToString(accOwner)
		require.NoError(t, err)
		withdrawAddr, err := app.AuthKeeper.AddressCodec().BytesToString(withdrawAcc)
		require.NoError(t, err)
		msg := &types.MsgWithdraw{
			Withdrawer: ownerAddr,
			ToAddress:  withdrawAddr,
			Denoms:     []string{"stake"},
		}
		err = s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NotNil(t, err)
	})

	// Update context time
	// After first period 500stake should be unlock
	ctx = ctx.WithHeaderInfo(header.Info{
		Time: currentTime.Add(time.Minute),
	})

	// Check if 500 stake is sendable now
	t.Run("ok - execute send message", func(t *testing.T) {
		msg := &types.MsgSend{
			Sender:    ownerAddrStr,
			ToAddress: addr,
			Amount:    sdk.Coins{sdk.NewCoin("stake", math.NewInt(500))},
		}
		err := s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NoError(t, err)

		balance := app.BankKeeper.GetBalance(ctx, randAcc, "stake")
		require.True(t, balance.Amount.Equal(math.NewInt(500)))
	})

	// Update context time
	// After second period 1000stake should be unlock
	ctx = ctx.WithHeaderInfo(header.Info{
		Time: currentTime.Add(time.Minute * 2),
	})

	t.Run("oke - execute withdraw message", func(t *testing.T) {
		ownerAddr, err := app.AuthKeeper.AddressCodec().BytesToString(accOwner)
		require.NoError(t, err)
		withdrawAddr, err := app.AuthKeeper.AddressCodec().BytesToString(withdrawAcc)
		require.NoError(t, err)
		msg := &types.MsgWithdraw{
			Withdrawer: ownerAddr,
			ToAddress:  withdrawAddr,
			Denoms:     []string{"stake"},
		}
		err = s.executeTx(ctx, msg, app, accountAddr, accOwner)
		require.NoError(t, err)

		// withdrawable amount should be
		// 1000stake - 500stake( above sent amt ) = 500stake
		balance := app.BankKeeper.GetBalance(ctx, withdrawAcc, "stake")
		require.True(t, balance.Amount.Equal(math.NewInt(500)))
	})

	// Fund acc since we withdraw all the funds
	s.fundAccount(app, ctx, accountAddr, sdk.Coins{sdk.NewCoin("stake", math.NewInt(100))})

	t.Run("ok - execute delegate message", func(t *testing.T) {
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
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
	})
}

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
		vals, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err)
		val := vals[0]
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
