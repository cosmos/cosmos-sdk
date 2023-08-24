package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/simapp"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestCancelUnbondingDelegation(t *testing.T) {
	// setup the app
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)
	bondDenom := app.StakingKeeper.BondDenom(ctx)

	// set the not bonded pool module account
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 5)

	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	moduleBalance := app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, sdk.NewInt64Coin(bondDenom, startTokens.Int64()), moduleBalance)

	// accounts
	delAddrs := simtestutil.AddTestAddrsIncremental(app.BankKeeper, app.StakingKeeper, ctx, 2, sdk.NewInt(10000))
	validators := app.StakingKeeper.GetValidators(ctx, 10)
	require.Equal(t, len(validators), 1)

	validatorAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(t, err)
	delegatorAddr := delAddrs[0]

	// setting the ubd entry
	unbondingAmount := sdk.NewInt64Coin(app.StakingKeeper.BondDenom(ctx), 5)
	ubd := types.NewUnbondingDelegation(
		delegatorAddr, validatorAddr, 10,
		ctx.BlockTime().Add(time.Minute*10),
		unbondingAmount.Amount,
		0,
	)

	// set and retrieve a record
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	resUnbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	require.True(t, found)
	require.Equal(t, ubd, resUnbond)

	testCases := []struct {
		Name      string
		ExceptErr bool
		req       types.MsgCancelUnbondingDelegation
	}{
		{
			Name:      "invalid height",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(4)),
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid coin",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin("dump_coin", sdk.NewInt(4)),
				CreationHeight:   0,
			},
		},
		{
			Name:      "validator not exists",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: sdk.ValAddress(sdk.AccAddress("asdsad")).String(),
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid delegator address",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: "invalid_delegator_addrtess",
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid amount",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Add(sdk.NewInt64Coin(bondDenom, 10)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1))),
				CreationHeight:   10,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := msgServer.CancelUnbondingDelegation(ctx, &testCase.req)
			if testCase.ExceptErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				balanceForNotBondedPool := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(notBondedPool.GetAddress()), bondDenom)
				require.Equal(t, balanceForNotBondedPool, moduleBalance.Sub(testCase.req.Amount))
				moduleBalance = moduleBalance.Sub(testCase.req.Amount)
			}
		})
	}
}

func TestTokenizeSharesAndRedeemTokens(t *testing.T) {
	_, app, ctx := createTestInput(t)

	liquidStakingCapStrict := sdk.ZeroDec()
	liquidStakingCapConservative := sdk.MustNewDecFromStr("0.8")
	liquidStakingCapDisabled := sdk.OneDec()

	validatorBondStrict := sdk.OneDec()
	validatorBondConservative := sdk.NewDec(10)
	validatorBondDisabled := sdk.NewDec(-1)

	testCases := []struct {
		name                          string
		vestingAmount                 sdk.Int
		delegationAmount              sdk.Int
		tokenizeShareAmount           sdk.Int
		redeemAmount                  sdk.Int
		targetVestingDelAfterShare    sdk.Int
		targetVestingDelAfterRedeem   sdk.Int
		globalLiquidStakingCap        sdk.Dec
		slashFactor                   sdk.Dec
		validatorLiquidStakingCap     sdk.Dec
		validatorBondFactor           sdk.Dec
		validatorBondDelegation       bool
		validatorBondDelegatorIndex   int
		delegatorIsLSTP               bool
		expTokenizeErr                bool
		expRedeemErr                  bool
		prevAccountDelegationExists   bool
		recordAccountDelegationExists bool
	}{
		{
			name:                          "full amount tokenize and redeem",
			vestingAmount:                 sdk.NewInt(0),
			delegationAmount:              app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                  app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:                   sdk.ZeroDec(),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			validatorBondFactor:           validatorBondDisabled,
			validatorBondDelegation:       false,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   false,
			recordAccountDelegationExists: false,
		},
		{
			name:                          "full amount tokenize and partial redeem",
			vestingAmount:                 sdk.NewInt(0),
			delegationAmount:              app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                   sdk.ZeroDec(),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			validatorBondFactor:           validatorBondDisabled,
			validatorBondDelegation:       false,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   false,
			recordAccountDelegationExists: true,
		},
		{
			name:                          "partial amount tokenize and full redeem",
			vestingAmount:                 sdk.NewInt(0),
			delegationAmount:              app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                   sdk.ZeroDec(),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			validatorBondFactor:           validatorBondDisabled,
			validatorBondDelegation:       false,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   true,
			recordAccountDelegationExists: false,
		},
		{
			name:                          "tokenize and redeem with slash",
			vestingAmount:                 sdk.NewInt(0),
			delegationAmount:              app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                   sdk.MustNewDecFromStr("0.1"),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			validatorBondFactor:           validatorBondDisabled,
			validatorBondDelegation:       false,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   false,
			recordAccountDelegationExists: true,
		},
		{
			name:                      "over tokenize",
			vestingAmount:             sdk.NewInt(0),
			delegationAmount:          app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:       app.StakingKeeper.TokensFromConsensusPower(ctx, 30),
			redeemAmount:              app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:               sdk.ZeroDec(),
			globalLiquidStakingCap:    liquidStakingCapDisabled,
			validatorLiquidStakingCap: liquidStakingCapDisabled,
			validatorBondFactor:       validatorBondDisabled,
			validatorBondDelegation:   false,
			expTokenizeErr:            true,
			expRedeemErr:              false,
		},
		{
			name:                      "over redeem",
			vestingAmount:             sdk.NewInt(0),
			delegationAmount:          app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:       app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:              app.StakingKeeper.TokensFromConsensusPower(ctx, 40),
			slashFactor:               sdk.ZeroDec(),
			globalLiquidStakingCap:    liquidStakingCapDisabled,
			validatorLiquidStakingCap: liquidStakingCapDisabled,
			validatorBondFactor:       validatorBondDisabled,
			validatorBondDelegation:   false,
			expTokenizeErr:            false,
			expRedeemErr:              true,
		},
		{
			name:                        "vesting account tokenize share failure",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     false,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "vesting account tokenize share success",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     false,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "try tokenize share for a validator-bond delegation",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondConservative,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 1,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "strict validator-bond - tokenization fails",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondStrict,
			validatorBondDelegation:     false,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "conservative validator-bond - successful tokenization",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondConservative,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "strict global liquid staking cap - tokenization fails",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapStrict,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "conservative global liquid staking cap - successful tokenization",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapConservative,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "strict validator liquid staking cap - tokenization fails",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapStrict,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "conservative validator liquid staking cap - successful tokenization",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapConservative,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "all caps set conservatively - successful tokenize share",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapConservative,
			validatorLiquidStakingCap:   liquidStakingCapConservative,
			validatorBondFactor:         validatorBondConservative,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "delegator is a liquid staking provider - accounting should not update",
			vestingAmount:               sdk.ZeroInt(),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 sdk.ZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapConservative,
			validatorLiquidStakingCap:   liquidStakingCapConservative,
			validatorBondFactor:         validatorBondConservative,
			delegatorIsLSTP:             true,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, app, ctx = createTestInput(t)
			addrs := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 2, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
			addrAcc1, addrAcc2 := addrs[0], addrs[1]
			addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

			// Create ICA module account
			icaAccountAddress := createICAAccount(app, ctx)

			// Fund module account
			delegationCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), tc.delegationAmount)
			err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(delegationCoin))
			require.NoError(t, err)
			err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, icaAccountAddress, sdk.NewCoins(delegationCoin))
			require.NoError(t, err)

			// set the delegator address depending on whether the delegator should be a liquid staking provider
			delegatorAccount := addrAcc2
			if tc.delegatorIsLSTP {
				delegatorAccount = icaAccountAddress
			}

			// set validator bond factor and global liquid staking cap
			params := app.StakingKeeper.GetParams(ctx)
			params.ValidatorBondFactor = tc.validatorBondFactor
			params.GlobalLiquidStakingCap = tc.globalLiquidStakingCap
			params.ValidatorLiquidStakingCap = tc.validatorLiquidStakingCap
			app.StakingKeeper.SetParams(ctx, params)

			// set the total liquid staked tokens
			app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, sdk.ZeroInt())

			if !tc.vestingAmount.IsZero() {
				// create vesting account
				pubkey := secp256k1.GenPrivKey().PubKey()
				baseAcc := authtypes.NewBaseAccount(addrAcc2, pubkey, 0, 0)
				initialVesting := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tc.vestingAmount))
				baseVestingWithCoins := vestingtypes.NewBaseVestingAccount(baseAcc, initialVesting, ctx.BlockTime().Unix()+86400*365)
				delayedVestingAccount := vestingtypes.NewDelayedVestingAccountRaw(baseVestingWithCoins)
				app.AccountKeeper.SetAccount(ctx, delayedVestingAccount)
			}

			pubKeys := simtestutil.CreateTestPubKeys(2)
			pk1, pk2 := pubKeys[0], pubKeys[1]

			// Create Validators and Delegation
			val1 := stakingtestutil.NewValidator(t, addrVal1, pk1)
			val1.Status = types.Bonded
			app.StakingKeeper.SetValidator(ctx, val1)
			app.StakingKeeper.SetValidatorByPowerIndex(ctx, val1)
			err = app.StakingKeeper.SetValidatorByConsAddr(ctx, val1)
			require.NoError(t, err)

			val2 := stakingtestutil.NewValidator(t, addrVal2, pk2)
			val2.Status = types.Bonded
			app.StakingKeeper.SetValidator(ctx, val2)
			app.StakingKeeper.SetValidatorByPowerIndex(ctx, val2)
			err = app.StakingKeeper.SetValidatorByConsAddr(ctx, val2)
			require.NoError(t, err)

			// Delegate from both the main delegator as well as a random account so there is a
			// non-zero delegation after redemption
			err = delegateCoinsFromAccount(ctx, app, delegatorAccount, tc.delegationAmount, val1)
			require.NoError(t, err)

			// apply TM updates
			applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)

			_, found := app.StakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			require.True(t, found, "delegation not found after delegate")

			lastRecordID := app.StakingKeeper.GetLastTokenizeShareRecordID(ctx)
			oldValidator, found := app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)

			msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)
			if tc.validatorBondDelegation {
				err := delegateCoinsFromAccount(ctx, app, addrs[tc.validatorBondDelegatorIndex], tc.delegationAmount, val1)
				require.NoError(t, err)
				_, err = msgServer.ValidatorBond(sdk.WrapSDKContext(ctx), &types.MsgValidatorBond{
					DelegatorAddress: addrs[tc.validatorBondDelegatorIndex].String(),
					ValidatorAddress: addrVal1.String(),
				})
				require.NoError(t, err)
			}

			resp, err := msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
				DelegatorAddress:    delegatorAccount.String(),
				ValidatorAddress:    addrVal1.String(),
				Amount:              sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), tc.tokenizeShareAmount),
				TokenizedShareOwner: delegatorAccount.String(),
			})
			if tc.expTokenizeErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// check last record id increase
			require.Equal(t, lastRecordID+1, app.StakingKeeper.GetLastTokenizeShareRecordID(ctx))

			// ensure validator's total tokens is consistent
			newValidator, found := app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)
			require.Equal(t, oldValidator.Tokens, newValidator.Tokens)

			// if the delegator was not a provider, check that the total liquid staked and validator liquid shares increased
			totalLiquidTokensAfterTokenization := app.StakingKeeper.GetTotalLiquidStakedTokens(ctx)
			validatorLiquidSharesAfterTokenization := newValidator.LiquidShares
			if !tc.delegatorIsLSTP {
				require.Equal(t, tc.tokenizeShareAmount.String(), totalLiquidTokensAfterTokenization.String(), "total liquid tokens after tokenization")
				require.Equal(t, tc.tokenizeShareAmount.String(), validatorLiquidSharesAfterTokenization.TruncateInt().String(), "validator liquid shares after tokenization")
			} else {
				require.True(t, totalLiquidTokensAfterTokenization.IsZero(), "zero liquid tokens after tokenization")
				require.True(t, validatorLiquidSharesAfterTokenization.IsZero(), "zero liquid validator shares after tokenization")
			}

			if tc.vestingAmount.IsPositive() {
				acc := app.AccountKeeper.GetAccount(ctx, addrAcc2)
				vestingAcc := acc.(vesting.VestingAccount)
				require.Equal(t, vestingAcc.GetDelegatedVesting().AmountOf(app.StakingKeeper.BondDenom(ctx)).String(), tc.targetVestingDelAfterShare.String())
			}

			if tc.prevAccountDelegationExists {
				_, found = app.StakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
				require.True(t, found, "delegation found after partial tokenize share")
			} else {
				_, found = app.StakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
				require.False(t, found, "delegation found after full tokenize share")
			}

			shareToken := app.BankKeeper.GetBalance(ctx, delegatorAccount, resp.Amount.Denom)
			require.Equal(t, resp.Amount, shareToken)
			_, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found, true, "validator not found")

			records := app.StakingKeeper.GetAllTokenizeShareRecords(ctx)
			require.Len(t, records, 1)
			delegation, found := app.StakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
			require.True(t, found, "delegation not found from tokenize share module account after tokenize share")

			// slash before redeem
			slashedTokens := sdk.ZeroInt()
			redeemedShares := tc.redeemAmount
			redeemedTokens := tc.redeemAmount
			if tc.slashFactor.IsPositive() {
				consAddr, err := val1.GetConsAddr()
				require.NoError(t, err)
				ctx = ctx.WithBlockHeight(100)
				val1, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
				require.True(t, found)
				power := app.StakingKeeper.TokensToConsensusPower(ctx, val1.Tokens)
				app.StakingKeeper.Slash(ctx, consAddr, 10, power, tc.slashFactor)
				slashedTokens = sdk.NewDecFromInt(val1.Tokens).Mul(tc.slashFactor).TruncateInt()

				val1, _ := app.StakingKeeper.GetValidator(ctx, addrVal1)
				redeemedTokens = val1.TokensFromShares(sdk.NewDecFromInt(redeemedShares)).TruncateInt()
			}

			// get deletagor balance and delegation
			bondDenomAmountBefore := app.BankKeeper.GetBalance(ctx, delegatorAccount, app.StakingKeeper.BondDenom(ctx))
			val1, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)
			delegation, found = app.StakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			if !found {
				delegation = types.Delegation{Shares: sdk.ZeroDec()}
			}
			delAmountBefore := val1.TokensFromShares(delegation.Shares)
			oldValidator, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)

			_, err = msgServer.RedeemTokensForShares(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensForShares{
				DelegatorAddress: delegatorAccount.String(),
				Amount:           sdk.NewCoin(resp.Amount.Denom, tc.redeemAmount),
			})
			if tc.expRedeemErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// ensure validator's total tokens is consistent
			newValidator, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)
			require.Equal(t, oldValidator.Tokens, newValidator.Tokens)

			// if the delegator was not a liuqid staking provider, check that the total liquid staked
			// and liquid shares decreased
			totalLiquidTokensAfterRedemption := app.StakingKeeper.GetTotalLiquidStakedTokens(ctx)
			validatorLiquidSharesAfterRedemption := newValidator.LiquidShares
			expectedLiquidTokens := totalLiquidTokensAfterTokenization.Sub(redeemedTokens).Sub(slashedTokens)
			expectedLiquidShares := validatorLiquidSharesAfterTokenization.Sub(sdk.NewDecFromInt(redeemedShares))
			if !tc.delegatorIsLSTP {
				require.Equal(t, expectedLiquidTokens.String(), totalLiquidTokensAfterRedemption.String(), "total liquid tokens after redemption")
				require.Equal(t, expectedLiquidShares.String(), validatorLiquidSharesAfterRedemption.String(), "validator liquid shares after tokenization")
			} else {
				require.True(t, totalLiquidTokensAfterRedemption.IsZero(), "zero liquid tokens after redemption")
				require.True(t, validatorLiquidSharesAfterRedemption.IsZero(), "zero liquid validator shares after redemption")
			}

			if tc.vestingAmount.IsPositive() {
				acc := app.AccountKeeper.GetAccount(ctx, addrAcc2)
				vestingAcc := acc.(vesting.VestingAccount)
				require.Equal(t, vestingAcc.GetDelegatedVesting().AmountOf(app.StakingKeeper.BondDenom(ctx)).String(), tc.targetVestingDelAfterRedeem.String())
			}

			expectedDelegatedShares := sdk.NewDecFromInt(tc.delegationAmount.Sub(tc.tokenizeShareAmount).Add(tc.redeemAmount))
			delegation, found = app.StakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			require.True(t, found, "delegation not found after redeem tokens")
			require.Equal(t, delegatorAccount.String(), delegation.DelegatorAddress)
			require.Equal(t, addrVal1.String(), delegation.ValidatorAddress)
			require.Equal(t, expectedDelegatedShares, delegation.Shares, "delegation shares after redeem")

			// check delegator balance is not changed
			bondDenomAmountAfter := app.BankKeeper.GetBalance(ctx, delegatorAccount, app.StakingKeeper.BondDenom(ctx))
			require.Equal(t, bondDenomAmountAfter.Amount.String(), bondDenomAmountBefore.Amount.String())

			// get delegation amount is changed correctly
			val1, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)
			delegation, found = app.StakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			if !found {
				delegation = types.Delegation{Shares: sdk.ZeroDec()}
			}
			delAmountAfter := val1.TokensFromShares(delegation.Shares)
			require.Equal(t, delAmountAfter.String(), delAmountBefore.Add(sdk.NewDecFromInt(tc.redeemAmount).Mul(sdk.OneDec().Sub(tc.slashFactor))).String())

			shareToken = app.BankKeeper.GetBalance(ctx, delegatorAccount, resp.Amount.Denom)
			require.Equal(t, shareToken.Amount.String(), tc.tokenizeShareAmount.Sub(tc.redeemAmount).String())
			_, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found, true, "validator not found")

			if tc.recordAccountDelegationExists {
				_, found = app.StakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
				require.True(t, found, "delegation not found from tokenize share module account after redeem partial amount")

				records = app.StakingKeeper.GetAllTokenizeShareRecords(ctx)
				require.Len(t, records, 1)
			} else {
				_, found = app.StakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
				require.False(t, found, "delegation found from tokenize share module account after redeem full amount")

				records = app.StakingKeeper.GetAllTokenizeShareRecords(ctx)
				require.Len(t, records, 0)
			}
		})
	}
}

// Helper function to setup a delegator and validator for the Tokenize/Redeem conversion tests
func setupTestTokenizeAndRedeemConversion(
	t *testing.T,
	app *simapp.SimApp,
	ctx sdk.Context,
) (delAddress sdk.AccAddress, valAddress sdk.ValAddress) {
	addresses := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 2, sdk.NewInt(1_000_000))
	pubKeys := simtestutil.CreateTestPubKeys(1)

	delegatorAddress := addresses[0]
	validatorAddress := sdk.ValAddress(addresses[1])

	validator := stakingtestutil.NewValidator(t, validatorAddress, pubKeys[0])
	validator.DelegatorShares = sdk.NewDec(1_000_000)
	validator.Tokens = sdk.NewInt(1_000_000)
	validator.LiquidShares = sdk.NewDec(0)
	validator.Status = types.Bonded

	app.StakingKeeper.SetValidator(ctx, validator)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)

	return delegatorAddress, validatorAddress
}

// Simulate a slash by decrementing the validator's tokens
// We'll do this in a way such that the exchange rate is not an even integer
// and the shares associated with a delegation will have a long decimal
func simulateSlashWithImprecision(t *testing.T, app *simapp.SimApp, ctx sdk.Context, valAddress sdk.ValAddress) {
	validator, found := app.StakingKeeper.GetValidator(ctx, valAddress)
	require.True(t, found)

	slashMagnitude := sdk.MustNewDecFromStr("0.1111111111")
	slashTokens := sdk.NewDecFromInt(validator.Tokens).Mul(slashMagnitude).TruncateInt()
	validator.Tokens = validator.Tokens.Sub(slashTokens)

	app.StakingKeeper.SetValidator(ctx, validator)
}

// Tests the conversion from tokenization and redemption from the following scenario:
// Slash -> Delegate -> Tokenize -> Redeem
// Note, in this example, there 2 tokens are lost during the decimal to int conversion
// during the unbonding step within tokenization and redemption
func TestTokenizeAndRedeemConversion_SlashBeforeDelegation(t *testing.T) {
	_, app, ctx := createTestInput(t)
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	delegatorAddress, validatorAddress := setupTestTokenizeAndRedeemConversion(t, app, ctx)

	// slash the validator
	simulateSlashWithImprecision(t, app, ctx, validatorAddress)
	validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddress)
	require.True(t, found)

	// Delegate and confirm the delegation record was created
	delegateAmount := sdk.NewInt(1000)
	delegateCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegateAmount)
	_, err := msgServer.Delegate(sdk.WrapSDKContext(ctx), &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err, "no error expected when delegating")

	delegation, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.True(t, found, "delegation should have been found")

	// Tokenize the full delegation amount
	_, err = msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              delegateCoin,
		TokenizedShareOwner: delegatorAddress.String(),
	})
	require.NoError(t, err, "no error expected when tokenizing")

	// Confirm the number of shareTokens equals the number of shares truncated
	// Note: 1 token is lost during unbonding due to rounding
	shareDenom := validatorAddress.String() + "/1"
	shareToken := app.BankKeeper.GetBalance(ctx, delegatorAddress, shareDenom)
	expectedShareTokens := delegation.Shares.TruncateInt().Int64() - 1 // 1 token was lost during unbonding
	require.Equal(t, expectedShareTokens, shareToken.Amount.Int64(), "share token amount")

	// Redeem the share tokens
	_, err = msgServer.RedeemTokensForShares(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
		Amount:           shareToken,
	})
	require.NoError(t, err, "no error expected when redeeming")

	// Confirm (almost) the full delegation was recovered - minus the 2 tokens from the precision error
	// (1 occurs during tokenization, and 1 occurs during redemption)
	newDelegation, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.True(t, found)

	endDelegationTokens := validator.TokensFromShares(newDelegation.Shares).TruncateInt().Int64()
	expectedDelegationTokens := delegateAmount.Int64() - 2
	require.Equal(t, expectedDelegationTokens, endDelegationTokens, "final delegation tokens")
}

// Tests the conversion from tokenization and redemption from the following scenario:
// Delegate -> Slash -> Tokenize -> Redeem
// Note, in this example, there 1 token lost during the decimal to int conversion
// during the unbonding step within tokenization
func TestTokenizeAndRedeemConversion_SlashBeforeTokenization(t *testing.T) {
	_, app, ctx := createTestInput(t)
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	delegatorAddress, validatorAddress := setupTestTokenizeAndRedeemConversion(t, app, ctx)

	// Delegate and confirm the delegation record was created
	delegateAmount := sdk.NewInt(1000)
	delegateCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegateAmount)
	_, err := msgServer.Delegate(sdk.WrapSDKContext(ctx), &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err, "no error expected when delegating")

	_, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.True(t, found, "delegation should have been found")

	// slash the validator
	simulateSlashWithImprecision(t, app, ctx, validatorAddress)
	validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddress)
	require.True(t, found)

	// Tokenize the new amount after the slash
	delegationAmountAfterSlash := validator.TokensFromShares(sdk.NewDecFromInt(delegateAmount)).TruncateInt()
	tokenizationCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegationAmountAfterSlash)

	_, err = msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              tokenizationCoin,
		TokenizedShareOwner: delegatorAddress.String(),
	})
	require.NoError(t, err, "no error expected when tokenizing")

	// The number of share tokens should line up with the **new** number of shares associated
	// with the original delegated amount
	// Note: 1 token is lost during unbonding due to rounding
	shareDenom := validatorAddress.String() + "/1"
	shareToken := app.BankKeeper.GetBalance(ctx, delegatorAddress, shareDenom)
	expectedShareTokens, err := validator.SharesFromTokens(tokenizationCoin.Amount)
	require.Equal(t, expectedShareTokens.TruncateInt().Int64()-1, shareToken.Amount.Int64(), "share token amount")

	// // Redeem the share tokens
	_, err = msgServer.RedeemTokensForShares(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
		Amount:           shareToken,
	})
	require.NoError(t, err, "no error expected when redeeming")

	// Confirm the full tokenization amount was recovered - minus the 1 token from the precision error
	newDelegation, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.True(t, found)

	endDelegationTokens := validator.TokensFromShares(newDelegation.Shares).TruncateInt().Int64()
	expectedDelegationTokens := delegationAmountAfterSlash.Int64() - 1
	require.Equal(t, expectedDelegationTokens, endDelegationTokens, "final delegation tokens")
}

// Tests the conversion from tokenization and redemption from the following scenario:
// Delegate -> Tokenize -> Slash -> Redeem
// Note, in this example, there 1 token lost during the decimal to int conversion
// during the unbonding step within redemption
func TestTokenizeAndRedeemConversion_SlashBeforeRedemptino(t *testing.T) {
	_, app, ctx := createTestInput(t)
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	delegatorAddress, validatorAddress := setupTestTokenizeAndRedeemConversion(t, app, ctx)

	// Delegate and confirm the delegation record was created
	delegateAmount := sdk.NewInt(1000)
	delegateCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegateAmount)
	_, err := msgServer.Delegate(sdk.WrapSDKContext(ctx), &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err, "no error expected when delegating")

	_, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.True(t, found, "delegation should have been found")

	// Tokenize the full delegation amount
	_, err = msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              delegateCoin,
		TokenizedShareOwner: delegatorAddress.String(),
	})
	require.NoError(t, err, "no error expected when tokenizing")

	// The number of share tokens should line up 1:1 with the number of issued shares
	// Since the validator has not been slashed, the shares also line up 1;1
	// with the original delegation amount
	shareDenom := validatorAddress.String() + "/1"
	shareToken := app.BankKeeper.GetBalance(ctx, delegatorAddress, shareDenom)
	expectedShareTokens := delegateAmount
	require.Equal(t, expectedShareTokens.Int64(), shareToken.Amount.Int64(), "share token amount")

	// slash the validator
	simulateSlashWithImprecision(t, app, ctx, validatorAddress)
	validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddress)
	require.True(t, found)

	// Redeem the share tokens
	_, err = msgServer.RedeemTokensForShares(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
		Amount:           shareToken,
	})
	require.NoError(t, err, "no error expected when redeeming")

	// Confirm the original delegation, minus the slash, was recovered
	// There's an additional 1 token lost from precision error during unbonding
	delegationAmountAfterSlash := validator.TokensFromShares(sdk.NewDecFromInt(delegateAmount)).TruncateInt().Int64()
	newDelegation, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.True(t, found)

	endDelegationTokens := validator.TokensFromShares(newDelegation.Shares).TruncateInt().Int64()
	require.Equal(t, delegationAmountAfterSlash-1, endDelegationTokens, "final delegation tokens")
}

func TestTransferTokenizeShareRecord(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrs := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 3, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1, addrAcc2, valAcc := addrs[0], addrs[1], addrs[2]
	addrVal := sdk.ValAddress(valAcc)

	pubKeys := simtestutil.CreateTestPubKeys(1)
	pk := pubKeys[0]

	val := stakingtestutil.NewValidator(t, addrVal, pk)
	app.StakingKeeper.SetValidator(ctx, val)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, val)

	// apply TM updates
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)

	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	err := app.StakingKeeper.AddTokenizeShareRecord(ctx, types.TokenizeShareRecord{
		Id:            1,
		Owner:         addrAcc1.String(),
		ModuleAccount: "module_account",
		Validator:     val.String(),
	})
	require.NoError(t, err)

	_, err = msgServer.TransferTokenizeShareRecord(sdk.WrapSDKContext(ctx), &types.MsgTransferTokenizeShareRecord{
		TokenizeShareRecordId: 1,
		Sender:                addrAcc1.String(),
		NewOwner:              addrAcc2.String(),
	})
	require.NoError(t, err)

	record, err := app.StakingKeeper.GetTokenizeShareRecord(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, record.Owner, addrAcc2.String())

	records := app.StakingKeeper.GetTokenizeShareRecordsByOwner(ctx, addrAcc1)
	require.Len(t, records, 0)
	records = app.StakingKeeper.GetTokenizeShareRecordsByOwner(ctx, addrAcc2)
	require.Len(t, records, 1)
}

func TestValidatorBond(t *testing.T) {
	_, app, ctx := createTestInput(t)

	testCases := []struct {
		name                 string
		createValidator      bool
		createDelegation     bool
		alreadyValidatorBond bool
		delegatorIsLSTP      bool
		expectedErr          error
	}{
		{
			name:                 "successful validator bond",
			createValidator:      true,
			createDelegation:     true,
			alreadyValidatorBond: false,
			delegatorIsLSTP:      false,
		},
		{
			name:                 "successful with existing validator bond",
			createValidator:      true,
			createDelegation:     true,
			alreadyValidatorBond: true,
			delegatorIsLSTP:      false,
		},
		{
			name:                 "validator does not not exist",
			createValidator:      false,
			createDelegation:     false,
			alreadyValidatorBond: false,
			delegatorIsLSTP:      false,
			expectedErr:          types.ErrNoValidatorFound,
		},
		{
			name:                 "delegation not exist case",
			createValidator:      true,
			createDelegation:     false,
			alreadyValidatorBond: false,
			delegatorIsLSTP:      false,
			expectedErr:          types.ErrNoDelegation,
		},
		{
			name:                 "delegator is a liquid staking provider",
			createValidator:      true,
			createDelegation:     true,
			alreadyValidatorBond: false,
			delegatorIsLSTP:      true,
			expectedErr:          types.ErrValidatorBondNotAllowedFromModuleAccount,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, app, ctx = createTestInput(t)

			pubKeys := simtestutil.CreateTestPubKeys(2)
			validatorPubKey := pubKeys[0]
			delegatorPubKey := pubKeys[1]

			delegatorAddress := sdk.AccAddress(delegatorPubKey.Address())
			validatorAddress := sdk.ValAddress(validatorPubKey.Address())
			icaAccountAddress := createICAAccount(app, ctx)

			// Set the delegator address to either be a user account or an ICA account depending on the test case
			if tc.delegatorIsLSTP {
				delegatorAddress = icaAccountAddress
			}

			// Fund the delegator
			delegationAmount := app.StakingKeeper.TokensFromConsensusPower(ctx, 20)
			coins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegationAmount))

			err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
			require.NoError(t, err, "no error expected when minting")

			err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delegatorAddress, coins)
			require.NoError(t, err, "no error expected when funding account")

			// Create Validator and delegation
			if tc.createValidator {
				validator := stakingtestutil.NewValidator(t, validatorAddress, validatorPubKey)
				validator.Status = types.Bonded
				app.StakingKeeper.SetValidator(ctx, validator)
				app.StakingKeeper.SetValidatorByPowerIndex(ctx, validator)
				err = app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)
				require.NoError(t, err)

				// Optionally create the delegation, depending on the test case
				if tc.createDelegation {
					_, err = app.StakingKeeper.Delegate(ctx, delegatorAddress, delegationAmount, types.Unbonded, validator, true)
					require.NoError(t, err, "no error expected when delegating")

					// Optionally, convert the delegation into a validator bond
					if tc.alreadyValidatorBond {
						delegation, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
						require.True(t, found, "delegation should have been found")

						delegation.ValidatorBond = true
						app.StakingKeeper.SetDelegation(ctx, delegation)
					}
				}
			}

			// Call ValidatorBond
			msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)
			_, err = msgServer.ValidatorBond(sdk.WrapSDKContext(ctx), &types.MsgValidatorBond{
				DelegatorAddress: delegatorAddress.String(),
				ValidatorAddress: validatorAddress.String(),
			})

			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err, "no error expected from validator bond transaction")

				// check validator bond true
				delegation, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
				require.True(t, found, "delegation should have been found after validator bond")
				require.True(t, delegation.ValidatorBond, "delegation should be marked as a validator bond")

				// check total validator bond shares
				validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddress)
				require.True(t, found, "validator should have been found after validator bond")

				if tc.alreadyValidatorBond {
					require.True(t, validator.ValidatorBondShares.IsZero(), "validator total shares should still be zero")
				} else {
					require.Equal(t, delegation.Shares.String(), validator.ValidatorBondShares.String(),
						"validator total shares should have increased")
				}
			}
		})
	}
}

func TestChangeValidatorBond(t *testing.T) {
	_, app, ctx := createTestInput(t)
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	checkValidatorBondShares := func(validatorAddress sdk.ValAddress, expectedShares sdk.Int) {
		validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddress)
		require.True(t, found, "validator should have been found")
		require.Equal(t, expectedShares.Int64(), validator.ValidatorBondShares.TruncateInt64(), "validator bond shares")
	}

	// Create a delegator and 3 validators
	addresses := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 4, sdk.NewInt(1_000_000))
	pubKeys := simtestutil.CreateTestPubKeys(4)

	validatorAPubKey := pubKeys[1]
	validatorBPubKey := pubKeys[2]
	validatorCPubKey := pubKeys[3]

	delegatorAddress := addresses[0]
	validatorAAddress := sdk.ValAddress(validatorAPubKey.Address())
	validatorBAddress := sdk.ValAddress(validatorBPubKey.Address())
	validatorCAddress := sdk.ValAddress(validatorCPubKey.Address())

	validatorA := stakingtestutil.NewValidator(t, validatorAAddress, validatorAPubKey)
	validatorB := stakingtestutil.NewValidator(t, validatorBAddress, validatorBPubKey)
	validatorC := stakingtestutil.NewValidator(t, validatorCAddress, validatorCPubKey)

	validatorA.Tokens = sdk.NewInt(1_000_000)
	validatorB.Tokens = sdk.NewInt(1_000_000)
	validatorC.Tokens = sdk.NewInt(1_000_000)
	validatorA.DelegatorShares = sdk.NewDec(1_000_000)
	validatorB.DelegatorShares = sdk.NewDec(1_000_000)
	validatorC.DelegatorShares = sdk.NewDec(1_000_000)

	app.StakingKeeper.SetValidator(ctx, validatorA)
	app.StakingKeeper.SetValidator(ctx, validatorB)
	app.StakingKeeper.SetValidator(ctx, validatorC)

	// The test will go through Delegate/Redelegate/Undelegate messages with the following
	delegation1Amount := sdk.NewInt(1000)
	delegation2Amount := sdk.NewInt(1000)
	redelegateAmount := sdk.NewInt(500)
	undelegateAmount := sdk.NewInt(500)

	delegate1Coin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegation1Amount)
	delegate2Coin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegation2Amount)
	redelegateCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), redelegateAmount)
	undelegateCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), undelegateAmount)

	// Delegate to validator's A and C - validator bond shares should not change
	_, err := msgServer.Delegate(sdk.WrapSDKContext(ctx), &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           delegate1Coin,
	})
	require.NoError(t, err, "no error expected during first delegation")

	_, err = msgServer.Delegate(sdk.WrapSDKContext(ctx), &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorCAddress.String(),
		Amount:           delegate1Coin,
	})
	require.NoError(t, err, "no error expected during first delegation")

	checkValidatorBondShares(validatorAAddress, sdk.ZeroInt())
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())
	checkValidatorBondShares(validatorCAddress, sdk.ZeroInt())

	// Flag the the delegations to validator A and C validator bond's
	// Their bond shares should increase
	_, err = msgServer.ValidatorBond(sdk.WrapSDKContext(ctx), &types.MsgValidatorBond{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
	})
	require.NoError(t, err, "no error expected during validator bond")

	_, err = msgServer.ValidatorBond(sdk.WrapSDKContext(ctx), &types.MsgValidatorBond{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorCAddress.String(),
	})
	require.NoError(t, err, "no error expected during validator bond")

	checkValidatorBondShares(validatorAAddress, delegation1Amount)
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())
	checkValidatorBondShares(validatorCAddress, delegation1Amount)

	// Delegate more to validator A - it should increase the validator bond shares
	_, err = msgServer.Delegate(sdk.WrapSDKContext(ctx), &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           delegate2Coin,
	})
	require.NoError(t, err, "no error expected during second delegation")

	checkValidatorBondShares(validatorAAddress, delegation1Amount.Add(delegation2Amount))
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())
	checkValidatorBondShares(validatorCAddress, delegation1Amount)

	// Redelegate partially from A to B (where A is a validator bond and B is not)
	// It should remove the bond shares from A, and B's validator bond shares should not change
	_, err = msgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), &types.MsgBeginRedelegate{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorSrcAddress: validatorAAddress.String(),
		ValidatorDstAddress: validatorBAddress.String(),
		Amount:              redelegateCoin,
	})
	require.NoError(t, err, "no error expected during redelegation")

	expectedBondSharesA := delegation1Amount.Add(delegation2Amount).Sub(redelegateAmount)
	checkValidatorBondShares(validatorAAddress, expectedBondSharesA)
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())
	checkValidatorBondShares(validatorCAddress, delegation1Amount)

	// Now redelegate from B to C (where B is not a validator bond, but C is)
	// Validator B's bond shares should remain at zero, but C's bond shares should increase
	_, err = msgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), &types.MsgBeginRedelegate{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorSrcAddress: validatorBAddress.String(),
		ValidatorDstAddress: validatorCAddress.String(),
		Amount:              redelegateCoin,
	})
	require.NoError(t, err, "no error expected during redelegation")

	checkValidatorBondShares(validatorAAddress, expectedBondSharesA)
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())
	checkValidatorBondShares(validatorCAddress, delegation1Amount.Add(redelegateAmount))

	// Redelegate partially from A to C (where C is a validator bond delegation)
	// It should remove the bond shares from A, and increase the bond shares on validator C
	_, err = msgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), &types.MsgBeginRedelegate{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorSrcAddress: validatorAAddress.String(),
		ValidatorDstAddress: validatorCAddress.String(),
		Amount:              redelegateCoin,
	})
	require.NoError(t, err, "no error expected during redelegation")

	expectedBondSharesA = expectedBondSharesA.Sub(redelegateAmount)
	expectedBondSharesC := delegation1Amount.Add(redelegateAmount).Add(redelegateAmount)
	checkValidatorBondShares(validatorAAddress, expectedBondSharesA)
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())
	checkValidatorBondShares(validatorCAddress, expectedBondSharesC)

	// Undelegate from validator A - it should remove shares
	_, err = msgServer.Undelegate(sdk.WrapSDKContext(ctx), &types.MsgUndelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           undelegateCoin,
	})
	require.NoError(t, err, "no error expected during undelegation")

	expectedBondSharesA = expectedBondSharesA.Sub(undelegateAmount)
	checkValidatorBondShares(validatorAAddress, expectedBondSharesA)
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())
	checkValidatorBondShares(validatorCAddress, expectedBondSharesC)
}

func TestEnableDisableTokenizeShares(t *testing.T) {
	_, app, ctx := createTestInput(t)
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	// Create a delegator and validator
	stakeAmount := sdk.NewInt(1000)
	stakeToken := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), stakeAmount)

	addresses := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 2, stakeAmount)
	delegatorAddress := addresses[0]

	pubKeys := simtestutil.CreateTestPubKeys(1)
	validatorAddress := sdk.ValAddress(addresses[1])
	validator := stakingtestutil.NewValidator(t, validatorAddress, pubKeys[0])

	validator.DelegatorShares = sdk.NewDec(1_000_000)
	validator.Tokens = sdk.NewInt(1_000_000)
	validator.Status = types.Bonded
	app.StakingKeeper.SetValidator(ctx, validator)

	// Fix block time and set unbonding period to 1 day
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(blockTime)

	unbondingPeriod := time.Hour * 24
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = unbondingPeriod
	app.StakingKeeper.SetParams(ctx, params)
	unlockTime := blockTime.Add(unbondingPeriod)

	// Build test messages (some of which will be reused)
	delegateMsg := types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           stakeToken,
	}
	tokenizeMsg := types.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              stakeToken,
		TokenizedShareOwner: delegatorAddress.String(),
	}
	redeemMsg := types.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
	}
	disableMsg := types.MsgDisableTokenizeShares{
		DelegatorAddress: delegatorAddress.String(),
	}
	enableMsg := types.MsgEnableTokenizeShares{
		DelegatorAddress: delegatorAddress.String(),
	}

	// Delegate normally
	_, err := msgServer.Delegate(sdk.WrapSDKContext(ctx), &delegateMsg)
	require.NoError(t, err, "no error expected when delegating")

	// Tokenize shares - it should succeed
	_, err = msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &tokenizeMsg)
	require.NoError(t, err, "no error expected when tokenizing shares for the first time")

	liquidToken := app.BankKeeper.GetBalance(ctx, delegatorAddress, validatorAddress.String()+"/1")
	require.Equal(t, stakeAmount.Int64(), liquidToken.Amount.Int64(), "user received token after tokenizing share")

	// Redeem to remove all tokenized shares
	redeemMsg.Amount = liquidToken
	_, err = msgServer.RedeemTokensForShares(sdk.WrapSDKContext(ctx), &redeemMsg)
	require.NoError(t, err, "no error expected when redeeming")

	// Attempt to enable tokenizing shares when there is no lock in place, it should error
	_, err = msgServer.EnableTokenizeShares(sdk.WrapSDKContext(ctx), &enableMsg)
	require.ErrorIs(t, err, types.ErrTokenizeSharesAlreadyEnabledForAccount)

	// Attempt to disable when no lock is in place, it should succeed
	_, err = msgServer.DisableTokenizeShares(sdk.WrapSDKContext(ctx), &disableMsg)
	require.NoError(t, err, "no error expected when disabling tokenization")

	// Disabling again while the lock is already in place, should error
	_, err = msgServer.DisableTokenizeShares(sdk.WrapSDKContext(ctx), &disableMsg)
	require.ErrorIs(t, err, types.ErrTokenizeSharesAlreadyDisabledForAccount)

	// Attempt to tokenize, it should fail since tokenization is disabled
	_, err = msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &tokenizeMsg)
	require.ErrorIs(t, err, types.ErrTokenizeSharesDisabledForAccount)

	// Now enable tokenization
	_, err = msgServer.EnableTokenizeShares(sdk.WrapSDKContext(ctx), &enableMsg)
	require.NoError(t, err, "no error expected when enabling tokenization")

	// Attempt to tokenize again, it should still fail since the unbonding period has
	// not passed and the lock is still active
	_, err = msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &tokenizeMsg)
	require.ErrorIs(t, err, types.ErrTokenizeSharesDisabledForAccount)
	require.ErrorContains(t, err, fmt.Sprintf("tokenization will be allowed at %s",
		blockTime.Add(unbondingPeriod)))

	// Confirm the unlock is queued
	authorizations := app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, unlockTime)
	require.Equal(t, []string{delegatorAddress.String()}, authorizations.Addresses,
		"pending tokenize share authorizations")

	// Disable tokenization again - it should remove the pending record from the queue
	_, err = msgServer.DisableTokenizeShares(sdk.WrapSDKContext(ctx), &disableMsg)
	require.NoError(t, err, "no error expected when re-enabling tokenization")

	authorizations = app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, unlockTime)
	require.Empty(t, authorizations.Addresses, "there should be no pending authorizations in the queue")

	// Enable one more time
	_, err = msgServer.EnableTokenizeShares(sdk.WrapSDKContext(ctx), &enableMsg)
	require.NoError(t, err, "no error expected when enabling tokenization again")

	// Increment the block time by the unbonding period and remove the expired locks
	ctx = ctx.WithBlockTime(unlockTime)
	app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, ctx.BlockTime())

	// Attempt to tokenize again, it should succeed this time since the lock has expired
	_, err = msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &tokenizeMsg)
	require.NoError(t, err, "no error expected when tokenizing after lock has expired")
}

func TestUnbondValidator(t *testing.T) {
	_, app, ctx := createTestInput(t)
	addrs := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 2, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1 := addrs[0]
	addrVal1 := sdk.ValAddress(addrAcc1)

	pubKeys := simtestutil.CreateTestPubKeys(1)
	pk1 := pubKeys[0]

	// Create Validators and Delegation
	val1 := stakingtestutil.NewValidator(t, addrVal1, pk1)
	val1.Status = types.Bonded
	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, val1)
	err := app.StakingKeeper.SetValidatorByConsAddr(ctx, val1)
	require.NoError(t, err)

	// try unbonding not available validator
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)
	_, err = msgServer.UnbondValidator(sdk.WrapSDKContext(ctx), &types.MsgUnbondValidator{
		ValidatorAddress: sdk.ValAddress(addrs[1]).String(),
	})
	require.Error(t, err)

	// unbond validator
	_, err = msgServer.UnbondValidator(sdk.WrapSDKContext(ctx), &types.MsgUnbondValidator{
		ValidatorAddress: addrVal1.String(),
	})
	require.NoError(t, err)

	// check if validator is jailed
	validator, found := app.StakingKeeper.GetValidator(ctx, addrVal1)
	require.True(t, found)
	require.True(t, validator.Jailed)
}

func TestChangeValidatorBond(t *testing.T) {
	_, app, ctx := createTestInput(t)
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	checkValidatorBondShares := func(validatorAddress sdk.ValAddress, expectedShares sdk.Int) {
		validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddress)
		require.True(t, found, "validator should have been found")
		require.Equal(t, expectedShares.Int64(), validator.ValidatorBondShares.TruncateInt64(), "validator bond shares")
	}

	// Create a delegator and 2 validators
	addresses := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 3, sdk.NewInt(1_000_000))
	pubKeys := simtestutil.CreateTestPubKeys(3)

	validatorAPubKey := pubKeys[1]
	validatorBPubKey := pubKeys[2]

	delegatorAddress := addresses[0]
	validatorAAddress := sdk.ValAddress(validatorAPubKey.Address())
	validatorBAddress := sdk.ValAddress(validatorBPubKey.Address())

	validatorA := stakingtestutil.NewValidator(t, validatorAAddress, validatorAPubKey)
	validatorB := stakingtestutil.NewValidator(t, validatorBAddress, validatorBPubKey)

	validatorA.Tokens = sdk.NewInt(1_000_000)
	validatorB.Tokens = sdk.NewInt(1_000_000)
	validatorA.DelegatorShares = sdk.NewDec(1_000_000)
	validatorB.DelegatorShares = sdk.NewDec(1_000_000)

	app.StakingKeeper.SetValidator(ctx, validatorA)
	app.StakingKeeper.SetValidator(ctx, validatorB)

	// The test will go through Delegate/Redelegate/Undelegate messages with the following
	delegation1Amount := sdk.NewInt(1000)
	delegation2Amount := sdk.NewInt(1000)
	redelegateAmount := sdk.NewInt(500)
	undelegateAmount := sdk.NewInt(500)

	delegate1Coin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegation1Amount)
	delegate2Coin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegation2Amount)
	redelegateCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), redelegateAmount)
	undelegateCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), undelegateAmount)

	// Delegate to validator A - validator bond shares should not change
	_, err := msgServer.Delegate(sdk.WrapSDKContext(ctx), &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           delegate1Coin,
	})
	require.NoError(t, err, "no error expected during first delegation")

	checkValidatorBondShares(validatorAAddress, sdk.ZeroInt())
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())

	// Flag the delegation as a validator bond
	_, err = msgServer.ValidatorBond(sdk.WrapSDKContext(ctx), &types.MsgValidatorBond{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
	})
	require.NoError(t, err, "no error expected during validator bond")

	checkValidatorBondShares(validatorAAddress, delegation1Amount)
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())

	// Delegate more - it should increase the validator bond shares
	_, err = msgServer.Delegate(sdk.WrapSDKContext(ctx), &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           delegate2Coin,
	})
	require.NoError(t, err, "no error expected during second delegation")

	checkValidatorBondShares(validatorAAddress, delegation1Amount.Add(delegation2Amount))
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())

	// Redelegate partially from A to B - it should remove the bond shares from the source validator
	_, err = msgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), &types.MsgBeginRedelegate{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorSrcAddress: validatorAAddress.String(),
		ValidatorDstAddress: validatorBAddress.String(),
		Amount:              redelegateCoin,
	})
	require.NoError(t, err, "no error expected during redelegation")

	checkValidatorBondShares(validatorAAddress, delegation1Amount.Add(delegation2Amount).Sub(redelegateAmount))
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())

	// Undelegate from validator A - it should have removed the shares
	_, err = msgServer.Undelegate(sdk.WrapSDKContext(ctx), &types.MsgUndelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           undelegateCoin,
	})
	require.NoError(t, err, "no error expected during undelegation")

	expectedBondShares := delegation1Amount.Add(delegation2Amount).Sub(redelegateAmount).Sub(undelegateAmount)
	checkValidatorBondShares(validatorAAddress, expectedBondShares)
	checkValidatorBondShares(validatorBAddress, sdk.ZeroInt())
}

// TestICADelegateUndelegate tests that an ICA account can undelegate
// sequentially right after delegating.
func TestICADelegateUndelegate(t *testing.T) {
	_, app, ctx := createTestInput(t)
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	// Create a delegator and validator (the delegator will be an ICA account)
	delegateAmount := sdk.NewInt(1000)
	delegateCoin := sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delegateAmount)
	icaAccountAddress := createICAAccount(app, ctx)

	// Fund ICA account
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(delegateCoin))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, icaAccountAddress, sdk.NewCoins(delegateCoin))
	require.NoError(t, err)

	addresses := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 1, sdk.NewInt(0))
	pubKeys := simtestutil.CreateTestPubKeys(1)
	validatorAddress := sdk.ValAddress(addresses[0])
	validator := stakingtestutil.NewValidator(t, validatorAddress, pubKeys[0])

	validator.DelegatorShares = sdk.NewDec(1_000_000)
	validator.Tokens = sdk.NewInt(1_000_000)
	validator.LiquidShares = sdk.NewDec(0)
	app.StakingKeeper.SetValidator(ctx, validator)

	delegateMsg := types.MsgDelegate{
		DelegatorAddress: icaAccountAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	}

	undelegateMsg := types.MsgUndelegate{
		DelegatorAddress: icaAccountAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	}

	// Delegate normally
	_, err = msgServer.Delegate(sdk.WrapSDKContext(ctx), &delegateMsg)
	require.NoError(t, err, "no error expected when delegating")

	// Confirm delegation record
	_, found := app.StakingKeeper.GetDelegation(ctx, icaAccountAddress, validatorAddress)
	require.True(t, found, "delegation should have been found")

	// Confirm liquid staking totals were incremented
	expectedTotalLiquidStaked := delegateAmount.Int64()
	actualTotalLiquidStaked := app.StakingKeeper.GetTotalLiquidStakedTokens(ctx).Int64()
	require.Equal(t, expectedTotalLiquidStaked, actualTotalLiquidStaked, "total liquid staked tokens after delegation")

	validator, found = app.StakingKeeper.GetValidator(ctx, validatorAddress)
	require.True(t, found, "validator should have been found")
	require.Equal(t, sdk.NewDecFromInt(delegateAmount), validator.LiquidShares, "validator liquid shares after delegation")

	// Try to undelegate
	_, err = msgServer.Undelegate(sdk.WrapSDKContext(ctx), &undelegateMsg)
	require.NoError(t, err, "no error expected when sequentially undelegating")

	// Confirm delegation record was removed
	_, found = app.StakingKeeper.GetDelegation(ctx, icaAccountAddress, validatorAddress)
	require.False(t, found, "delegation not have been found")

	// Confirm liquid staking totals were decremented
	actualTotalLiquidStaked = app.StakingKeeper.GetTotalLiquidStakedTokens(ctx).Int64()
	require.Zero(t, actualTotalLiquidStaked, "total liquid staked tokens after undelegation")

	validator, found = app.StakingKeeper.GetValidator(ctx, validatorAddress)
	require.True(t, found, "validator should have been found")
	require.Equal(t, sdk.ZeroDec(), validator.LiquidShares, "validator liquid shares after undelegation")
}

func TestCancelUnbondingDelegation(t *testing.T) {
	_, app, ctx := createTestInput(t)
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	bondDenom := app.StakingKeeper.BondDenom(ctx)

	// set the not bonded pool module account
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 5)
	startCoin := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), startTokens))

	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), startCoin))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	moduleBalance := app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, sdk.NewInt64Coin(bondDenom, startTokens.Int64()), moduleBalance)

	// create a  validator
	validatorPubKey := simtestutil.CreateTestPubKeys(1)[0]
	validatorAddr := sdk.ValAddress(validatorPubKey.Address())

	validator := stakingtestutil.NewValidator(t, validatorAddr, validatorPubKey)
	validator.Tokens = startTokens
	validator.DelegatorShares = sdk.NewDecFromInt(startTokens)
	validator.Status = types.Bonded
	app.StakingKeeper.SetValidator(ctx, validator)

	// create a delegator
	delAddrs := simtestutil.AddTestAddrsIncremental(app.BankKeeper, app.StakingKeeper, ctx, 2, sdk.NewInt(10000))
	delegatorAddr := delAddrs[0]

	// setting the ubd entry
	unbondingAmount := sdk.NewInt64Coin(app.StakingKeeper.BondDenom(ctx), 5)
	ubd := types.NewUnbondingDelegation(
		delegatorAddr, validatorAddr, 10,
		ctx.BlockTime().Add(time.Minute*10),
		unbondingAmount.Amount,
		1,
	)

	// set and retrieve a record
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	resUnbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	require.True(t, found)
	require.Equal(t, ubd, resUnbond)

	testCases := []struct {
		Name      string
		ExceptErr bool
		req       types.MsgCancelUnbondingDelegation
	}{
		{
			Name:      "invalid height",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(4)),
				CreationHeight:   0,
			},
		},
		{
			Name:      "validator not exists",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: sdk.ValAddress(sdk.AccAddress("asdsad")).String(),
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid delegator address",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: "invalid_delegator_addrtess",
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid amount",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Add(sdk.NewInt64Coin(bondDenom, 10)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1))),
				CreationHeight:   10,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := msgServer.CancelUnbondingDelegation(sdk.WrapSDKContext(ctx), &testCase.req)
			if testCase.ExceptErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				balanceForNotBondedPool := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(notBondedPool.GetAddress()), bondDenom)
				require.Equal(t, balanceForNotBondedPool, moduleBalance.Sub(testCase.req.Amount))
				moduleBalance = moduleBalance.Sub(testCase.req.Amount)
			}
		})
	}
}
