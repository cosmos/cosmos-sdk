package keeper_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"
	stakingtypes "cosmossdk.io/x/staking/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var TestProposal = getTestProposal()

func getTestProposal() []sdk.Msg {
	legacyProposalMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Title", "description"), authtypes.NewModuleAddress(types.ModuleName).String())
	if err != nil {
		panic(err)
	}
	testProposal := v1beta1.NewTextProposal("Proposal", "testing proposal")
	legacyProposalMsg2, err := v1.NewLegacyContent(testProposal, authtypes.NewModuleAddress(types.ModuleName).String())
	if err != nil {
		panic(err)
	}

	return []sdk.Msg{
		legacyProposalMsg,
		legacyProposalMsg2,
	}
}

func createValidators(t *testing.T, f *fixture, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress) {
	t.Helper()
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.ctx, 5, math.NewInt(30000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	pks := simtestutil.CreateTestPubKeys(5)

	val1, err := stakingtypes.NewValidator(valAddrs[0].String(), pks[0], stakingtypes.Description{})
	assert.NilError(t, err)
	val2, err := stakingtypes.NewValidator(valAddrs[1].String(), pks[1], stakingtypes.Description{})
	assert.NilError(t, err)
	val3, err := stakingtypes.NewValidator(valAddrs[2].String(), pks[2], stakingtypes.Description{})
	assert.NilError(t, err)

	assert.NilError(t, f.stakingKeeper.SetValidator(f.ctx, val1))
	assert.NilError(t, f.stakingKeeper.SetValidator(f.ctx, val2))
	assert.NilError(t, f.stakingKeeper.SetValidator(f.ctx, val3))
	assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.ctx, val1))
	assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.ctx, val2))
	assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.ctx, val3))
	assert.NilError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.ctx, val1))
	assert.NilError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.ctx, val2))
	assert.NilError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.ctx, val3))

	for _, addr := range addrs {
		f.accountKeeper.SetAccount(f.ctx, f.accountKeeper.NewAccountWithAddress(f.ctx, addr))
	}

	_, _ = f.stakingKeeper.Delegate(f.ctx, addrs[0], f.stakingKeeper.TokensFromConsensusPower(f.ctx, powers[0]), stakingtypes.Unbonded, val1, true)
	_, _ = f.stakingKeeper.Delegate(f.ctx, addrs[1], f.stakingKeeper.TokensFromConsensusPower(f.ctx, powers[1]), stakingtypes.Unbonded, val2, true)
	_, _ = f.stakingKeeper.Delegate(f.ctx, addrs[2], f.stakingKeeper.TokensFromConsensusPower(f.ctx, powers[2]), stakingtypes.Unbonded, val3, true)

	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	return addrs, valAddrs
}
