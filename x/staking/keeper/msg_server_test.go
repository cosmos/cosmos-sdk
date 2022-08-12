package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simapp "github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	sdkstaking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestTokenizeSharesAndRedeemTokens(t *testing.T) {
	_, app, ctx := createTestInput()

	testCases := []struct {
		name                          string
		vestingAmount                 sdk.Int
		delegationAmount              sdk.Int
		tokenizeShareAmount           sdk.Int
		redeemAmount                  sdk.Int
		targetVestingDelAfterShare    sdk.Int
		targetVestingDelAfterRedeem   sdk.Int
		slashFactor                   sdk.Dec
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
			slashFactor:                   sdk.NewDecWithPrec(10, 2),
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
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   true,
			recordAccountDelegationExists: false,
		},
		{
			name:                "over tokenize",
			vestingAmount:       sdk.NewInt(0),
			delegationAmount:    app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount: app.StakingKeeper.TokensFromConsensusPower(ctx, 30),
			redeemAmount:        app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:         sdk.ZeroDec(),
			expTokenizeErr:      true,
			expRedeemErr:        false,
		},
		{
			name:                "over redeem",
			vestingAmount:       sdk.NewInt(0),
			delegationAmount:    app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount: app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:        app.StakingKeeper.TokensFromConsensusPower(ctx, 40),
			slashFactor:         sdk.ZeroDec(),
			expTokenizeErr:      false,
			expRedeemErr:        true,
		},
		{
			name:                        "vesting account tokenize share failure",
			vestingAmount:               app.StakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                app.StakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:                 sdk.ZeroDec(),
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
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, app, ctx = createTestInput()
			addrs := simapp.AddTestAddrs(app, ctx, 2, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
			addrAcc1, addrAcc2 := addrs[0], addrs[1]
			addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

			if !tc.vestingAmount.IsZero() {
				// create vesting account
				pubkey := secp256k1.GenPrivKey().PubKey()
				baseAcc := authtypes.NewBaseAccount(addrAcc2, pubkey, 0, 0)
				initialVesting := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tc.vestingAmount))
				baseVestingWithCoins := vestingtypes.NewBaseVestingAccount(baseAcc, initialVesting, ctx.BlockTime().Unix()+86400*365)
				delayedVestingAccount := vestingtypes.NewDelayedVestingAccountRaw(baseVestingWithCoins)
				app.AccountKeeper.SetAccount(ctx, delayedVestingAccount)
			}

			pubKeys := simapp.CreateTestPubKeys(2)
			pk1, pk2 := pubKeys[0], pubKeys[1]

			// Create Validators and Delegation
			val1 := teststaking.NewValidator(t, addrVal1, pk1)
			val1.Status = sdkstaking.Bonded
			app.StakingKeeper.SetValidator(ctx, val1)
			app.StakingKeeper.SetValidatorByPowerIndex(ctx, val1)
			app.StakingKeeper.SetValidatorByConsAddr(ctx, val1)

			val2 := teststaking.NewValidator(t, addrVal2, pk2)
			val2.Status = sdkstaking.Bonded
			app.StakingKeeper.SetValidator(ctx, val2)
			app.StakingKeeper.SetValidatorByPowerIndex(ctx, val2)
			app.StakingKeeper.SetValidatorByConsAddr(ctx, val2)

			delTokens := tc.delegationAmount
			err := delegateCoinsFromAccount(ctx, app, addrAcc2, delTokens, val1)
			require.NoError(t, err)

			// apply TM updates
			applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)

			_, found := app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
			require.True(t, found, "delegation not found after delegate")

			lastRecordId := app.StakingKeeper.GetLastTokenizeShareRecordId(ctx)
			oldValidator, found := app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)

			msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)
			resp, err := msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
				DelegatorAddress:    addrAcc2.String(),
				ValidatorAddress:    addrVal1.String(),
				Amount:              sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), tc.tokenizeShareAmount),
				TokenizedShareOwner: addrAcc2.String(),
			})
			if tc.expTokenizeErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// check last record id increase
			require.Equal(t, lastRecordId+1, app.StakingKeeper.GetLastTokenizeShareRecordId(ctx))

			// ensure validator's total tokens is consistent
			newValidator, found := app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)
			require.Equal(t, oldValidator.Tokens, newValidator.Tokens)

			if tc.vestingAmount.IsPositive() {
				acc := app.AccountKeeper.GetAccount(ctx, addrAcc2)
				vestingAcc := acc.(vesting.VestingAccount)
				require.Equal(t, vestingAcc.GetDelegatedVesting().AmountOf(app.StakingKeeper.BondDenom(ctx)).String(), tc.targetVestingDelAfterShare.String())
			}

			if tc.prevAccountDelegationExists {
				_, found = app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
				require.True(t, found, "delegation found after partial tokenize share")
			} else {
				_, found = app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
				require.False(t, found, "delegation found after full tokenize share")
			}

			shareToken := app.BankKeeper.GetBalance(ctx, addrAcc2, resp.Amount.Denom)
			require.Equal(t, resp.Amount, shareToken)
			_, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found, true, "validator not found")

			records := app.StakingKeeper.GetAllTokenizeShareRecords(ctx)
			require.Len(t, records, 1)
			delegation, found := app.StakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
			require.True(t, found, "delegation not found from tokenize share module account after tokenize share")

			// slash before redeem
			if tc.slashFactor.IsPositive() {
				consAddr, err := val1.GetConsAddr()
				require.NoError(t, err)
				ctx = ctx.WithBlockHeight(100)
				val1, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
				require.True(t, found)
				power := app.StakingKeeper.TokensToConsensusPower(ctx, val1.Tokens)
				app.StakingKeeper.Slash(ctx, consAddr, 10, power, tc.slashFactor)
			}

			// get deletagor balance and delegation
			bondDenomAmountBefore := app.BankKeeper.GetBalance(ctx, addrAcc2, app.StakingKeeper.BondDenom(ctx))
			val1, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)
			delegation, found = app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
			if !found {
				delegation = types.Delegation{Shares: sdk.ZeroDec()}
			}
			delAmountBefore := val1.TokensFromShares(delegation.Shares)
			oldValidator, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)

			_, err = msgServer.RedeemTokens(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensforShares{
				DelegatorAddress: addrAcc2.String(),
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

			if tc.vestingAmount.IsPositive() {
				acc := app.AccountKeeper.GetAccount(ctx, addrAcc2)
				vestingAcc := acc.(vesting.VestingAccount)
				require.Equal(t, vestingAcc.GetDelegatedVesting().AmountOf(app.StakingKeeper.BondDenom(ctx)).String(), tc.targetVestingDelAfterRedeem.String())
			}

			delegation, found = app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
			require.True(t, found, "delegation not found after redeem tokens")
			require.Equal(t, delegation.DelegatorAddress, addrAcc2.String())
			require.Equal(t, delegation.ValidatorAddress, addrVal1.String())
			require.Equal(t, delegation.Shares, tc.delegationAmount.Sub(tc.tokenizeShareAmount).Add(tc.redeemAmount).ToDec())

			// check delegator balance is not changed
			bondDenomAmountAfter := app.BankKeeper.GetBalance(ctx, addrAcc2, app.StakingKeeper.BondDenom(ctx))
			require.Equal(t, bondDenomAmountAfter.Amount.String(), bondDenomAmountBefore.Amount.String())

			// get delegation amount is changed correctly
			val1, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
			require.True(t, found)
			delegation, found = app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
			if !found {
				delegation = types.Delegation{Shares: sdk.ZeroDec()}
			}
			delAmountAfter := val1.TokensFromShares(delegation.Shares)
			require.Equal(t, delAmountAfter.String(), delAmountBefore.Add(tc.redeemAmount.ToDec().Mul(sdk.OneDec().Sub(tc.slashFactor))).String())

			shareToken = app.BankKeeper.GetBalance(ctx, addrAcc2, resp.Amount.Denom)
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

func TestTransferTokenizeShareRecord(t *testing.T) {
	_, app, ctx := createTestInput()

	addrs := simapp.AddTestAddrs(app, ctx, 3, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1, addrAcc2, valAcc := addrs[0], addrs[1], addrs[2]
	addrVal := sdk.ValAddress(valAcc)

	pubKeys := simapp.CreateTestPubKeys(1)
	pk := pubKeys[0]

	val := teststaking.NewValidator(t, addrVal, pk)
	app.StakingKeeper.SetValidator(ctx, val)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, val)

	// apply TM updates
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)

	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	err := app.StakingKeeper.AddTokenizeShareRecord(ctx, types.TokenizeShareRecord{
		Id:              1,
		Owner:           addrAcc1.String(),
		ShareTokenDenom: "share_token_denom",
		ModuleAccount:   "module_account",
		Validator:       val.String(),
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
