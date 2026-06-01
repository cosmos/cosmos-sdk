package keeper_test

import (
	"cosmossdk.io/math"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestApplyAndReturnValidatorSetUpdatesWithKeyRotation exercises the EndBlocker
// path where the validator set update loop has already appended an update for
// the validator's old consensus key (because its power changed) and then a
// queued key rotation runs. The new key must end up at the validator's current
// power, not at a delta against the prior update.
func (s *KeeperTestSuite) TestApplyAndReturnValidatorSetUpdatesWithKeyRotation() {
	require := s.Require()

	powerReduction := s.stakingKeeper.PowerReduction(s.ctx)

	// set up a bonded validator with 10 consensus power. LastValidatorPower
	// is intentionally not seeded, so the main loop will append an update
	// for the old key with the validator's current power.
	oldPk := ed25519.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(oldPk.Address())
	v, err := stakingtypes.NewValidator(valAddr.String(), oldPk, stakingtypes.Description{Moniker: "v"})
	require.NoError(err)
	v.Status = stakingtypes.Bonded
	v.Tokens = sdk.TokensFromConsensusPower(10, powerReduction)
	v.DelegatorShares = math.LegacyNewDecFromInt(v.Tokens)
	require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))
	require.NoError(s.stakingKeeper.SetValidatorByConsAddr(s.ctx, v))
	require.NoError(s.stakingKeeper.SetNewValidatorByPowerIndex(s.ctx, v))

	// queue a key rotation for the same validator
	newPk := ed25519.GenPrivKey().PubKey()
	require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

	updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
	require.NoError(err)

	// expected list, in order:
	//   [0] old @ 10  (from the main loop discovering a power change)
	//   [1] old @ 0   (from the rotation retiring the old key)
	//   [2] new @ 10  (from the rotation instating the new key)
	require.Len(updates, 3)

	oldCmtPk, err := cryptocodec.ToCmtProtoPublicKey(oldPk)
	require.NoError(err)
	newCmtPk, err := cryptocodec.ToCmtProtoPublicKey(newPk)
	require.NoError(err)

	require.Equal(oldCmtPk, updates[0].PubKey)
	require.Equal(int64(10), updates[0].Power)

	require.Equal(oldCmtPk, updates[1].PubKey)
	require.Equal(int64(0), updates[1].Power)

	require.Equal(newCmtPk, updates[2].PubKey)
	require.Equal(int64(10), updates[2].Power)

	// simulate cometbft applying the updates in order, last write wins per
	// key. the final state must have the old key removed and the new key at
	// the validator's current power.
	finalPower := map[string]int64{}
	for _, u := range updates {
		bz, err := u.PubKey.Marshal()
		require.NoError(err)
		finalPower[string(bz)] = u.Power
	}

	oldBz, err := oldCmtPk.Marshal()
	require.NoError(err)
	newBz, err := newCmtPk.Marshal()
	require.NoError(err)

	require.Equal(int64(0), finalPower[string(oldBz)])
	require.Equal(int64(10), finalPower[string(newBz)])
}
