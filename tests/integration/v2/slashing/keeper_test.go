package slashing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/comet"
	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/x/slashing/testutil"
	slashingtypes "cosmossdk.io/x/slashing/types"
	stakingtestutil "cosmossdk.io/x/staking/testutil"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestUnJailNotBonded(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	p, err := f.stakingKeeper.Params.Get(f.ctx)
	assert.NilError(t, err)
	p.MaxValidators = 5
	assert.NilError(t, f.stakingKeeper.Params.Set(f.ctx, p))
	pks := simtestutil.CreateTestPubKeys(6)
	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)

	// create max (5) validators all with the same power
	for i := uint32(0); i < p.MaxValidators; i++ {
		addr, val := f.valAddrs[i], pks[i]
		acc := f.accountKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(addr))
		f.accountKeeper.SetAccount(f.ctx, acc)
		tstaking.CreateValidatorWithValPower(addr, val, 100, true)
	}

	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	newHeight := int64(f.app.LastBlockHeight()) + 1 + 1
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: newHeight})

	// create a 6th validator with less power than the cliff validator (won't be bonded)
	addr, val := f.valAddrs[5], pks[5]
	acc := f.accountKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(addr))
	f.accountKeeper.SetAccount(f.ctx, acc)
	amt := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 50)
	msg := tstaking.CreateValidatorMsg(addr, val, amt)
	msg.MinSelfDelegation = amt
	msg.Description = stakingtypes.Description{Moniker: "TestValidator"}
	res, err := tstaking.CreateValidatorWithMsg(f.ctx, msg)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)

	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	newHeight = integration.HeaderInfoFromContext(f.ctx).Height + 1
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: newHeight})

	tstaking.CheckValidator(addr, stakingtypes.Unbonded, false)

	// unbond below minimum self-delegation
	assert.Equal(t, p.BondDenom, tstaking.Denom)
	accAddr, err := f.accountKeeper.AddressCodec().BytesToString(addr)
	assert.NilError(t, err)
	valAddr, err := f.stakingKeeper.ValidatorAddressCodec().BytesToString(addr)
	require.NoError(t, err)
	tstaking.Undelegate(accAddr, valAddr, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 1), true)

	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	newHeight = integration.HeaderInfoFromContext(f.ctx).Height + 1
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: newHeight})

	// verify that validator is jailed
	tstaking.CheckValidator(addr, -1, true)

	// verify we cannot unjail (yet)
	msgUnjail := slashingtypes.MsgUnjail{
		ValidatorAddr: addr.String(),
	}
	_, err = f.app.RunMsg(
		t,
		f.ctx,
		func(ctx context.Context) (transaction.Msg, error) {
			res, err := f.slashingMsgServer.Unjail(ctx, &msgUnjail)
			return res, err
		},
		integration.WithAutomaticCommit(),
	)
	assert.ErrorContains(t, err, "cannot be unjailed")

	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	newHeight = integration.HeaderInfoFromContext(f.ctx).Height + 1
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: newHeight})
	// bond to meet minimum self-delegationa
	accAddr, err = f.accountKeeper.AddressCodec().BytesToString(addr)
	assert.NilError(t, err)
	valAddr, err = f.stakingKeeper.ValidatorAddressCodec().BytesToString(addr)
	assert.NilError(t, err)
	tstaking.DelegateWithPower(accAddr, valAddr, 1)

	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	newHeight = integration.HeaderInfoFromContext(f.ctx).Height + 1
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: newHeight})

	// verify we can immediately unjail
	_, err = f.app.RunMsg(
		t,
		f.ctx,
		func(ctx context.Context) (transaction.Msg, error) {
			res, err := f.slashingMsgServer.Unjail(ctx, &msgUnjail)
			return res, err
		},
		integration.WithAutomaticCommit(),
	)
	assert.NilError(t, err)

	tstaking.CheckValidator(addr, -1, false)
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func TestHandleNewValidator(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	require.NoError(t, err)

	bondPool := f.stakingKeeper.GetBondedPool(f.ctx)

	initialBondPoolBal := f.bankKeeper.GetBalance(f.ctx, bondPool.GetAddress(), bondDenom).Amount

	pks := simtestutil.CreateTestPubKeys(1)
	addr, valpubkey := f.valAddrs[0], pks[0]
	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)
	signedBlocksWindow, err := f.slashingKeeper.SignedBlocksWindow(f.ctx)
	assert.NilError(t, err)
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: signedBlocksWindow + 1})
	assert.NilError(t, f.slashingKeeper.AddrPubkeyRelation.Set(f.ctx, pks[0].Address(), pks[0]))

	consaddr, err := f.stakingKeeper.ConsensusAddressCodec().BytesToString(valpubkey.Address())
	assert.NilError(t, err)

	info := slashingtypes.NewValidatorSigningInfo(consaddr, integration.HeaderInfoFromContext(f.ctx).Height, time.Unix(0, 0), false, int64(0))
	assert.NilError(t, f.slashingKeeper.ValidatorSigningInfo.Set(f.ctx, sdk.ConsAddress(valpubkey.Address()), info))
	assert.Equal(t, signedBlocksWindow+1, info.StartHeight)

	// Validator created
	acc := f.accountKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(addr))
	f.accountKeeper.SetAccount(f.ctx, acc)
	amt := tstaking.CreateValidatorWithValPower(addr, valpubkey, 100, true)

	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)

	assert.DeepEqual(
		t, f.bankKeeper.GetAllBalances(f.ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(bondDenom, testutil.InitTokens.Sub(amt))),
	)

	val, err := f.stakingKeeper.Validator(f.ctx, addr)
	require.NoError(t, err)
	assert.DeepEqual(t, amt, val.GetBondedTokens())

	// Now a validator, for two blocks
	assert.NilError(t, f.slashingKeeper.HandleValidatorSignature(f.ctx, valpubkey.Address(), 100, comet.BlockIDFlagCommit))
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: signedBlocksWindow + 2})
	assert.NilError(t, f.slashingKeeper.HandleValidatorSignature(f.ctx, valpubkey.Address(), 100, comet.BlockIDFlagAbsent))

	info, found := f.slashingKeeper.ValidatorSigningInfo.Get(f.ctx, sdk.ConsAddress(valpubkey.Address()))
	assert.Assert(t, found)
	assert.Equal(t, signedBlocksWindow+1, info.StartHeight)
	assert.Equal(t, int64(1), info.MissedBlocksCounter)
	assert.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been jailed or slashed
	validator, _ := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, sdk.GetConsAddress(valpubkey))
	assert.Equal(t, sdk.Bonded, validator.GetStatus())
	expTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 100).Add(initialBondPoolBal)
	assert.Assert(t, expTokens.Equal(f.bankKeeper.GetBalance(f.ctx, bondPool.GetAddress(), bondDenom).Amount))
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyJailed(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	pks := simtestutil.CreateTestPubKeys(1)
	addr, val := f.valAddrs[0], pks[0]
	power := int64(100)
	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)

	err := f.slashingKeeper.AddrPubkeyRelation.Set(f.ctx, pks[0].Address(), pks[0])
	assert.NilError(t, err)

	consaddr, err := f.stakingKeeper.ConsensusAddressCodec().BytesToString(val.Address())
	assert.NilError(t, err)

	info := slashingtypes.NewValidatorSigningInfo(consaddr, integration.HeaderInfoFromContext(f.ctx).Height, time.Unix(0, 0), false, int64(0))
	assert.NilError(t, f.slashingKeeper.ValidatorSigningInfo.Set(f.ctx, sdk.ConsAddress(val.Address()), info))

	acc := f.accountKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(addr))
	f.accountKeeper.SetAccount(f.ctx, acc)

	amt := tstaking.CreateValidatorWithValPower(addr, val, power, true)

	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)

	signedBlocksWindow, err := f.slashingKeeper.SignedBlocksWindow(f.ctx)
	assert.NilError(t, err)

	// 1000 first blocks OK
	height := int64(0)
	for ; height < signedBlocksWindow; height++ {
		f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: height})
		err = f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), power, comet.BlockIDFlagCommit)
		assert.NilError(t, err)
	}

	minSignedPerWindow, err := f.slashingKeeper.MinSignedPerWindow(f.ctx)
	assert.NilError(t, err)

	// 501 blocks missed
	for ; height < signedBlocksWindow+(signedBlocksWindow-minSignedPerWindow)+1; height++ {
		f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: height})
		err = f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), power, comet.BlockIDFlagAbsent)
		assert.NilError(t, err)
	}

	// end block
	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)

	// validator should have been jailed and slashed
	validator, _ := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, sdk.GetConsAddress(val))
	assert.Equal(t, sdk.Unbonding, validator.GetStatus())

	// validator should have been slashed
	resultingTokens := amt.Sub(f.stakingKeeper.TokensFromConsensusPower(f.ctx, 1))
	assert.DeepEqual(t, resultingTokens, validator.GetTokens())

	// another block missed
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: height})
	assert.NilError(t, f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), power, comet.BlockIDFlagAbsent))

	// validator should not have been slashed twice
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, sdk.GetConsAddress(val))
	assert.DeepEqual(t, resultingTokens, validator.GetTokens())
}

// Test a validator dipping in and out of the validator set
// Ensure that missed blocks are tracked correctly and that
// the start height of the signing info is reset correctly
func TestValidatorDippingInAndOut(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	params, err := f.stakingKeeper.Params.Get(f.ctx)
	require.NoError(t, err)
	params.MaxValidators = 1
	err = f.stakingKeeper.Params.Set(f.ctx, params)
	assert.NilError(t, err)
	power := int64(100)

	pks := simtestutil.CreateTestPubKeys(3)
	simtestutil.AddTestAddrsFromPubKeys(f.bankKeeper, f.stakingKeeper, f.ctx, pks, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 200))
	for _, pk := range pks {
		acc := f.accountKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(pk.Address()))
		f.accountKeeper.SetAccount(f.ctx, acc)
	}

	addr, val := pks[0].Address(), pks[0]
	consAddr := sdk.ConsAddress(addr)
	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)
	valAddr := sdk.ValAddress(addr)

	assert.NilError(t, f.slashingKeeper.AddrPubkeyRelation.Set(f.ctx, pks[0].Address(), pks[0]))

	consaddrStr, err := f.stakingKeeper.ConsensusAddressCodec().BytesToString(addr)
	assert.NilError(t, err)

	info := slashingtypes.NewValidatorSigningInfo(consaddrStr, integration.HeaderInfoFromContext(f.ctx).Height, time.Unix(0, 0), false, int64(0))
	assert.NilError(t, f.slashingKeeper.ValidatorSigningInfo.Set(f.ctx, consAddr, info))

	tstaking.CreateValidatorWithValPower(valAddr, val, power, true)
	validatorUpdates, err := f.stakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)
	// validator updates length should be equal to 2 as we already have one default validator
	assert.Equal(t, 2, len(validatorUpdates))
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// 100 first blocks OK
	height := int64(0)
	for ; height < int64(100); height++ {
		f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: height})
		assert.NilError(t, f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), power, comet.BlockIDFlagCommit))
	}

	// kick first validator out of validator set
	tstaking.CreateValidatorWithValPower(sdk.ValAddress(pks[1].Address()), pks[1], power+1, true)
	validatorUpdates, err = f.stakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(validatorUpdates))
	tstaking.CheckValidator(sdk.ValAddress(pks[1].Address()), stakingtypes.Bonded, false)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, false)

	// 600 more blocks happened
	height += 600
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: height})

	// store this height as we don't have block height value in context
	startHeight := height
	// validator added back in
	accAddr, err := f.accountKeeper.AddressCodec().BytesToString(sdk.AccAddress(pks[2].Address()))
	assert.NilError(t, err)
	vAddr, err := f.stakingKeeper.ValidatorAddressCodec().BytesToString(valAddr)
	assert.NilError(t, err)
	tstaking.DelegateWithPower(accAddr, vAddr, 50)

	validatorUpdates, err = f.stakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(validatorUpdates))
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)
	newPower := power + 50

	// validator misses a block
	assert.NilError(t, f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), newPower, comet.BlockIDFlagAbsent))
	height++

	// shouldn't be jailed/kicked yet
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	signedBlocksWindow, err := f.slashingKeeper.SignedBlocksWindow(f.ctx)
	assert.NilError(t, err)

	// validator misses an additional 500 more blocks within the SignedBlockWindow (here 1000 blocks).
	latest := signedBlocksWindow + height
	// misses 500 blocks + within the signing windows i.e. 700-1700
	// validators misses all 1000 blocks of a SignedBlockWindows
	for ; height < latest+1; height++ {
		err = f.slashingKeeper.HandleValidatorSignature(integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: height}), val.Address(), newPower, comet.BlockIDFlagAbsent)
		assert.NilError(t, err)
	}

	// should now be jailed & kicked
	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, true)

	info = slashingtypes.NewValidatorSigningInfo(consaddrStr, startHeight, time.Unix(0, 0), false, int64(0))
	err = f.slashingKeeper.ValidatorSigningInfo.Set(f.ctx, consAddr, info)
	assert.NilError(t, err)

	// check all the signing information
	signInfo, found := f.slashingKeeper.ValidatorSigningInfo.Get(f.ctx, consAddr)
	assert.Assert(t, found)
	assert.Equal(t, int64(700), signInfo.StartHeight)
	assert.Equal(t, int64(0), signInfo.MissedBlocksCounter)
	assert.Equal(t, int64(0), signInfo.IndexOffset)

	// some blocks pass
	height = int64(5000)
	f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: height})

	info = slashingtypes.NewValidatorSigningInfo(consaddrStr, integration.HeaderInfoFromContext(f.ctx).Height, time.Unix(0, 0), false, int64(0))
	err = f.slashingKeeper.ValidatorSigningInfo.Set(f.ctx, consAddr, info)
	assert.NilError(t, err)

	// validator rejoins and starts signing again
	err = f.stakingKeeper.Unjail(f.ctx, consAddr)
	assert.NilError(t, err)
	err = f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), newPower, comet.BlockIDFlagCommit)
	assert.NilError(t, err)

	// validator should not be kicked since we reset counter/array when it was jailed
	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// check start height is correctly set
	signInfo, found = f.slashingKeeper.ValidatorSigningInfo.Get(f.ctx, consAddr)
	assert.Assert(t, found)
	assert.Equal(t, height, signInfo.StartHeight)

	minSignedPerWindow, err := f.slashingKeeper.MinSignedPerWindow(f.ctx)
	assert.NilError(t, err)

	// validator misses 501 blocks after SignedBlockWindow period (1000 blocks)
	latest = signedBlocksWindow + height
	for ; height < latest+minSignedPerWindow; height++ {
		f.ctx = integration.SetHeaderInfo(f.ctx, coreheader.Info{Height: height})
		err = f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), newPower, comet.BlockIDFlagAbsent)
		assert.NilError(t, err)
	}

	// validator should now be jailed & kicked
	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)

	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, true)
}
