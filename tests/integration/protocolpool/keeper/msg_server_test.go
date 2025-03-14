package keeper

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func TestMsgFundCommunityPool(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// check pool balance

	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(100))
	err := f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	assert.NilError(t, err)

	addr := sdk.AccAddress(PKS[0].Address())
	addr2 := sdk.AccAddress(PKS[1].Address())
	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))

	// fund the account by minting and sending amount from distribution module to addr
	err = f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, amount)
	assert.NilError(t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, addr, amount)
	assert.NilError(t, err)

	testCases := []struct {
		name      string
		msg       *protocolpooltypes.MsgFundCommunityPool
		expErr    bool
		expErrMsg string
	}{
		{
			name: "no depositor address",
			msg: &protocolpooltypes.MsgFundCommunityPool{
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				Depositor: emptyDelAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid depositor address",
		},
		{
			name: "invalid coin",
			msg: &protocolpooltypes.MsgFundCommunityPool{
				Amount:    sdk.Coins{sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin("stake", 10)},
				Depositor: addr.String(),
			},
			expErr:    true,
			expErrMsg: "10stake,10stake: invalid coins",
		},
		{
			name: "depositor address with no funds",
			msg: &protocolpooltypes.MsgFundCommunityPool{
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				Depositor: addr2.String(),
			},
			expErr:    true,
			expErrMsg: "insufficient funds",
		},
		{
			name: "valid message",
			msg: &protocolpooltypes.MsgFundCommunityPool{
				Amount:    amount,
				Depositor: addr.String(),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := protocolpooltypes.MsgFundCommunityPool{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// query the community pool funds
				pool, err := f.protocolPoolKeeper.GetCommunityPool(f.sdkCtx)
				assert.NilError(t, err)
				assert.Assert(t, pool.Equal(amount))
				assert.Assert(t, f.bankKeeper.GetAllBalances(f.sdkCtx, addr).Empty())
			}
		})
	}
}

func TestMsgUpdateParams(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	initParams := protocolpooltypes.Params{}
	assert.NilError(t, f.protocolPoolKeeper.Params.Set(f.sdkCtx, initParams))

	testCases := []struct {
		name      string
		msg       *protocolpooltypes.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			msg: &protocolpooltypes.MsgUpdateParams{
				Authority: "invalid",
				Params:    initParams,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "all good",
			msg: &protocolpooltypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: protocolpooltypes.Params{
					EnabledDistributionDenoms: []string{"stake"},
				},
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
				integration.WithAutomaticCommit(),
			)

			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := protocolpooltypes.MsgUpdateParams{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// query the params and verify it has been updated
				params, _ := f.protocolPoolKeeper.Params.Get(f.sdkCtx)
				assert.DeepEqual(t, tc.msg.Params, params)
			}
		})
	}
}

func TestMsgCommunityPoolSpend(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(100))
	amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
	err := f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, amount)
	assert.NilError(t, err)

	// now send the funds to the protocolpool module
	err = f.bankKeeper.SendCoinsFromModuleToModule(f.sdkCtx, distrtypes.ModuleName, protocolpooltypes.ProtocolPoolDistrAccount, amount)
	assert.NilError(t, err)

	// set the funds to be distributable
	err = f.protocolPoolKeeper.SetToDistribute(f.sdkCtx)
	assert.NilError(t, err)

	recipient := sdk.AccAddress("addr1")

	testCases := []struct {
		name      string
		msg       *protocolpooltypes.MsgCommunityPoolSpend
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			msg: &protocolpooltypes.MsgCommunityPoolSpend{
				Authority: "invalid",
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "invalid recipient",
			msg: &protocolpooltypes.MsgCommunityPoolSpend{
				Authority: f.protocolPoolKeeper.GetAuthority(),
				Recipient: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed",
		},
		{
			name: "valid message",
			msg: &protocolpooltypes.MsgCommunityPoolSpend{
				Authority: f.protocolPoolKeeper.GetAuthority(),
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := protocolpooltypes.MsgCommunityPoolSpend{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// query the community pool to verify it has been updated
				communityPool, err := f.protocolPoolKeeper.GetCommunityPool(f.sdkCtx)
				assert.NilError(t, err)
				newPool, negative := initialFeePool.SafeSub(sdk.NewDecCoinsFromCoins(tc.msg.Amount...))
				assert.Assert(t, negative == false)
				assert.DeepEqual(t, communityPool.CommunityPool, newPool)
			}
		})
	}
}
