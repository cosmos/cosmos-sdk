package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestGovernanceDelegate(t *testing.T) {
	assert := assert.New(t)
	govKeeper, accKeeper, bankKeeper, stakingKeeper, distrKeeper, _, ctx := setupGovKeeper(t, mockAccountKeeperExpectations)
	mocks := mocks{
		accKeeper:          accKeeper,
		bankKeeper:         bankKeeper,
		stakingKeeper:      stakingKeeper,
		distributionKeeper: distrKeeper,
	}
	s := newFixture(t, ctx, 2, 3, 2, govKeeper, mocks)
	// Setup the delegators
	s.delegate(s.delAddrs[0], s.valAddrs[0], 1)
	s.delegate(s.delAddrs[0], s.valAddrs[1], 2)
	s.delegate(s.delAddrs[1], s.valAddrs[0], 5)
	s.delegate(s.delAddrs[2], s.valAddrs[1], 8)

	// Delegate to active governor
	err := govKeeper.DelegateToGovernor(ctx, s.delAddrs[0], s.activeGovernors[0].GetAddress())
	assert.NoError(err)
	err = govKeeper.DelegateToGovernor(ctx, s.delAddrs[1], s.activeGovernors[0].GetAddress())
	assert.NoError(err)
	// Delegate to inactive governor
	err = govKeeper.DelegateToGovernor(ctx, s.delAddrs[2], s.inactiveGovernor.GetAddress())
	assert.NoError(err)

	// get governance delegations
	deleg1, err := govKeeper.GovernanceDelegations.Get(ctx, s.delAddrs[0])
	if assert.NoError(err, "deleg1 not found") {
		assert.Equal(deleg1.DelegatorAddress, s.delAddrs[0].String())
		assert.Equal(deleg1.GovernorAddress, s.activeGovernors[0].GovernorAddress)
	}
	deleg2, err := govKeeper.GovernanceDelegations.Get(ctx, s.delAddrs[1])
	if assert.NoError(err, "deleg2 not found") {
		assert.Equal(deleg2.DelegatorAddress, s.delAddrs[1].String())
		assert.Equal(deleg2.GovernorAddress, s.activeGovernors[0].GovernorAddress)
	}
	deleg3, err := govKeeper.GovernanceDelegations.Get(ctx, s.delAddrs[2])
	if assert.NoError(err, "deleg3 not found") {
		assert.Equal(deleg3.DelegatorAddress, s.delAddrs[2].String())
		assert.Equal(deleg3.GovernorAddress, s.inactiveGovernor.GovernorAddress)
	}

	// Assert RedelegateToGovernor
	err = govKeeper.RedelegateToGovernor(ctx, s.delAddrs[0], s.inactiveGovernor.GetAddress())
	assert.NoError(err)
	var allDelegs []*v1.GovernanceDelegation
	err = govKeeper.GovernanceDelegationsByGovernor.Walk(ctx, collections.NewPrefixedPairRange[types.GovernorAddress, sdk.AccAddress](s.activeGovernors[0].GetAddress()), func(_ collections.Pair[types.GovernorAddress, sdk.AccAddress], delegation v1.GovernanceDelegation) (stop bool, err error) {
		allDelegs = append(allDelegs, &delegation)
		return false, nil
	})
	assert.NoError(err)
	assert.ElementsMatch(allDelegs, []*v1.GovernanceDelegation{&deleg2})
	allDelegs = nil
	err = govKeeper.GovernanceDelegationsByGovernor.Walk(ctx, collections.NewPrefixedPairRange[types.GovernorAddress, sdk.AccAddress](s.inactiveGovernor.GetAddress()), func(_ collections.Pair[types.GovernorAddress, sdk.AccAddress], delegation v1.GovernanceDelegation) (stop bool, err error) {
		allDelegs = append(allDelegs, &delegation)
		return false, nil
	})
	assert.NoError(err)
	deleg1.GovernorAddress = s.inactiveGovernor.GovernorAddress
	assert.ElementsMatch(allDelegs, []*v1.GovernanceDelegation{&deleg1, &deleg3})
	valShare1, err := govKeeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(s.activeGovernors[0].GetAddress(), s.valAddrs[0]))
	if assert.NoError(err, "valShare1 not found") {
		assert.Equal(valShare1, v1.GovernorValShares{
			GovernorAddress:  s.activeGovernors[0].GovernorAddress,
			ValidatorAddress: s.valAddrs[0].String(),
			Shares:           math.LegacyNewDec(5),
		})
	}
	_, err = govKeeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(s.activeGovernors[0].GetAddress(), s.valAddrs[1]))
	assert.ErrorIs(err, collections.ErrNotFound, "valShare should be removed")
	valShare2, err := govKeeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(s.inactiveGovernor.GetAddress(), s.valAddrs[0]))
	if assert.NoError(err, "valShare2 not found") {
		assert.Equal(valShare2, v1.GovernorValShares{
			GovernorAddress:  s.inactiveGovernor.GovernorAddress,
			ValidatorAddress: s.valAddrs[0].String(),
			Shares:           math.LegacyNewDec(1),
		})
	}
	valShare3, err := govKeeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(s.inactiveGovernor.GetAddress(), s.valAddrs[1]))
	if assert.NoError(err, "valShare3 not found") {
		assert.Equal(valShare3, v1.GovernorValShares{
			GovernorAddress:  s.inactiveGovernor.GovernorAddress,
			ValidatorAddress: s.valAddrs[1].String(),
			Shares:           math.LegacyNewDec(10),
		})
	}

	// Assert UndelegateFromGovernor
	err = govKeeper.UndelegateFromGovernor(ctx, s.delAddrs[0])
	assert.NoError(err)
	allDelegs = nil
	err = govKeeper.GovernanceDelegationsByGovernor.Walk(ctx, collections.NewPrefixedPairRange[types.GovernorAddress, sdk.AccAddress](s.activeGovernors[0].GetAddress()), func(_ collections.Pair[types.GovernorAddress, sdk.AccAddress], delegation v1.GovernanceDelegation) (stop bool, err error) {
		allDelegs = append(allDelegs, &delegation)
		return false, nil
	})
	assert.NoError(err)
	assert.ElementsMatch(allDelegs, []*v1.GovernanceDelegation{&deleg2})
	allDelegs = nil
	err = govKeeper.GovernanceDelegationsByGovernor.Walk(ctx, collections.NewPrefixedPairRange[types.GovernorAddress, sdk.AccAddress](s.inactiveGovernor.GetAddress()), func(_ collections.Pair[types.GovernorAddress, sdk.AccAddress], delegation v1.GovernanceDelegation) (stop bool, err error) {
		allDelegs = append(allDelegs, &delegation)
		return false, nil
	})
	assert.NoError(err)
	deleg1.GovernorAddress = s.inactiveGovernor.GovernorAddress
	assert.ElementsMatch(allDelegs, []*v1.GovernanceDelegation{&deleg3})
	valShare1, err = govKeeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(s.activeGovernors[0].GetAddress(), s.valAddrs[0]))
	if assert.NoError(err, "valShare1 not found") {
		assert.Equal(valShare1, v1.GovernorValShares{
			GovernorAddress:  s.activeGovernors[0].GovernorAddress,
			ValidatorAddress: s.valAddrs[0].String(),
			Shares:           math.LegacyNewDec(5),
		})
	}
	_, err = govKeeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(s.activeGovernors[0].GetAddress(), s.valAddrs[1]))
	assert.ErrorIs(err, collections.ErrNotFound, "valShare should be removed")
	_, err = govKeeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(s.inactiveGovernor.GetAddress(), s.valAddrs[0]))
	assert.ErrorIs(err, collections.ErrNotFound, "valShare should be removed")
	valShare3, err = govKeeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(s.inactiveGovernor.GetAddress(), s.valAddrs[1]))
	if assert.NoError(err, "valShare3 not found") {
		assert.Equal(valShare3, v1.GovernorValShares{
			GovernorAddress:  s.inactiveGovernor.GovernorAddress,
			ValidatorAddress: s.valAddrs[1].String(),
			Shares:           math.LegacyNewDec(8),
		})
	}
}
