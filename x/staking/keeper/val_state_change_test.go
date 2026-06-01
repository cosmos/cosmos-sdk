package keeper_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestApplyAndReturnValidatorSetUpdatesWithKeyRotation exercises the EndBlocker
// path where a queued cons key rotation runs in the same block as a power
// change for the same validator. With the deferred-apply design, the rotation
// (old@0, new@power) pair is emitted first, and any subsequent main-loop emit
// for the validator is routed through the new cons key so CometBFT's set ends
// up at new@power after applying updates in order.
func (s *KeeperTestSuite) TestApplyAndReturnValidatorSetUpdatesWithKeyRotation() {
	require := s.Require()

	powerReduction := s.stakingKeeper.PowerReduction(s.ctx)

	// set up a bonded validator with 10 consensus power. LastValidatorPower
	// is intentionally not seeded, so the main loop will discover a power
	// change and append an update.
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
	//   [0] old @ 0   (rotation retiring the old key, emitted first)
	//   [1] new @ 10  (rotation instating the new key)
	//   [2] new @ 10  (main-loop power-change emit, routed through the new
	//                  cons key by effectiveConsKeyForUpdate because the
	//                  rotation applyHeight is within the emit window)
	require.Len(updates, 3)

	oldCmtPk, err := cryptocodec.ToCmtProtoPublicKey(oldPk)
	require.NoError(err)
	newCmtPk, err := cryptocodec.ToCmtProtoPublicKey(newPk)
	require.NoError(err)

	require.Equal(oldCmtPk, updates[0].PubKey)
	require.Equal(int64(0), updates[0].Power)

	require.Equal(newCmtPk, updates[1].PubKey)
	require.Equal(int64(10), updates[1].Power)

	require.Equal(newCmtPk, updates[2].PubKey)
	require.Equal(int64(10), updates[2].Power)

	// the main-loop emit must not reference the old cons key — that is the
	// regression the routing helper prevents.
	for i, u := range updates[1:] {
		require.NotEqual(oldCmtPk, u.PubKey, "post-rotation emit %d still references old cons key", i+1)
	}

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

func (s *KeeperTestSuite) TestApplyAndReturnValidatorSetUpdates_RotationEmitSequence() {
	require := s.Require()

	s.T().Run("baseline rotation only", func(t *testing.T) {
		s.SetupTest()
		const H = int64(10)
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()

		s.ctx = s.ctx.WithBlockHeight(H)
		valAddr := s.bondedValidatorAtPower(oldPk, 10)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		// H EndBlock: only the rotation pair is emitted. LastValidatorPower
		// equals the live power, so the main loop adds nothing.
		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Len(updates, 2)
		require.Equal(cmtPk(t, oldPk), updates[0].PubKey)
		require.Equal(int64(0), updates[0].Power)
		require.Equal(cmtPk(t, newPk), updates[1].PubKey)
		require.Equal(int64(10), updates[1].Power)

		// H+1 EndBlock: nothing happens; drain and emit passes both no-op.
		s.ctx = s.ctx.WithBlockHeight(H + 1)
		updates, err = s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Empty(updates)

		// H+2 EndBlock: drain swaps state, emit pass no-op.
		s.ctx = s.ctx.WithBlockHeight(H + 2)
		updates, err = s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Empty(updates)

		stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
		require.NoError(err)
		got, err := stored.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(newPk.Address()).Bytes(), got)
	})

	s.T().Run("jail at H emits rotation pair plus zero power on new key", func(t *testing.T) {
		s.SetupTest()
		// bonded -> unbonding transition during EndBlock moves tokens between
		// the bonded and not-bonded module accounts; stub the bank side so
		// the keeper logic can run end to end.
		s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		const H = int64(10)
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()

		s.ctx = s.ctx.WithBlockHeight(H)
		valAddr := s.bondedValidatorAtPower(oldPk, 10)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
		// jailing resolves via OLD cons addr — the deferred swap has not run.
		require.NoError(s.stakingKeeper.Jail(s.ctx, sdk.ConsAddress(oldPk.Address())))

		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Len(updates, 3)
		// rotation pair first
		require.Equal(cmtPk(t, oldPk), updates[0].PubKey)
		require.Equal(int64(0), updates[0].Power)
		require.Equal(cmtPk(t, newPk), updates[1].PubKey)
		require.Equal(int64(10), updates[1].Power)
		// jail emit, routed through the helper, references NEW
		require.Equal(cmtPk(t, newPk), updates[2].PubKey)
		require.Equal(int64(0), updates[2].Power)

		// CometBFT's view at H+2: validator absent under both keys.
		set := applyCometSet(t, updates)
		require.NotContains(set, string(mustMarshal(t, cmtPk(t, oldPk))))
		require.NotContains(set, string(mustMarshal(t, cmtPk(t, newPk))))

		_ = valAddr
	})

	s.T().Run("jail at H+1 emits zero power on new key", func(t *testing.T) {
		s.SetupTest()
		s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		const H = int64(10)
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()

		s.ctx = s.ctx.WithBlockHeight(H)
		valAddr := s.bondedValidatorAtPower(oldPk, 10)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		// H EndBlock: emit rotation pair, no state swap yet.
		hUpdates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Len(hUpdates, 2)

		// H+1: slashing routes jail via VoteInfo's OLD cons addr. The
		// byConsAddr index for OLD still resolves because the swap is deferred.
		s.ctx = s.ctx.WithBlockHeight(H + 1)
		require.NoError(s.stakingKeeper.Jail(s.ctx, sdk.ConsAddress(oldPk.Address())))
		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Len(updates, 1)
		// the jail emit must reference NEW so CometBFT's H+3 set (which has
		// the new key from H's rotation pair) sees the zero power update.
		require.Equal(cmtPk(t, newPk), updates[0].PubKey)
		require.Equal(int64(0), updates[0].Power)

		// H+2: drain runs; validator is already jailed, but the swap on the
		// operator record still applies. No new emits.
		s.ctx = s.ctx.WithBlockHeight(H + 2)
		updates, err = s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Empty(updates)
	})

	s.T().Run("power change at H+1 emits new key at the new power", func(t *testing.T) {
		s.SetupTest()
		const H = int64(10)
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()

		s.ctx = s.ctx.WithBlockHeight(H)
		valAddr := s.bondedValidatorAtPower(oldPk, 10)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		// H EndBlock: emit rotation pair.
		hUpdates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Len(hUpdates, 2)

		// H+1: validator's power moves from 10 to 25 (e.g. delegation).
		s.ctx = s.ctx.WithBlockHeight(H + 1)
		s.setValidatorPower(valAddr, 25)
		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Len(updates, 1)
		require.Equal(cmtPk(t, newPk), updates[0].PubKey) // routed via helper
		require.Equal(int64(25), updates[0].Power)
	})

	s.T().Run("power change at H emits rotation pair plus new key at new power", func(t *testing.T) {
		s.SetupTest()
		const H = int64(10)
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()

		s.ctx = s.ctx.WithBlockHeight(H)
		valAddr := s.bondedValidatorAtPower(oldPk, 10)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
		s.setValidatorPower(valAddr, 25)

		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Len(updates, 3)
		require.Equal(cmtPk(t, oldPk), updates[0].PubKey)
		require.Equal(int64(0), updates[0].Power)
		// rotation emit reads validator's current state (already power 25)
		require.Equal(cmtPk(t, newPk), updates[1].PubKey)
		require.Equal(int64(25), updates[1].Power)
		// main loop emit, routed via helper
		require.Equal(cmtPk(t, newPk), updates[2].PubKey)
		require.Equal(int64(25), updates[2].Power)

		// no post-rotation emit references OLD
		for i, u := range updates[1:] {
			require.NotEqual(cmtPk(t, oldPk), u.PubKey, "update %d still references old key", i+1)
		}
	})

	s.T().Run("two validators rotate at H", func(t *testing.T) {
		s.SetupTest()
		const H = int64(10)
		oldPk1 := ed25519.GenPrivKey().PubKey()
		newPk1 := ed25519.GenPrivKey().PubKey()
		oldPk2 := ed25519.GenPrivKey().PubKey()
		newPk2 := ed25519.GenPrivKey().PubKey()

		s.ctx = s.ctx.WithBlockHeight(H)
		valAddr1 := s.bondedValidatorAtPower(oldPk1, 10)
		valAddr2 := s.bondedValidatorAtPower(oldPk2, 10)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr1, oldPk1, newPk1))
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr2, oldPk2, newPk2))

		// H EndBlock: two rotation pairs (4 updates total). Iteration order
		// across the apply queue is deterministic but not asserted here;
		// applyCometSet collapses to the final per-key powers.
		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Len(updates, 4)
		set := applyCometSet(t, updates)
		require.Equal(int64(10), set[string(mustMarshal(t, cmtPk(t, newPk1)))])
		require.Equal(int64(10), set[string(mustMarshal(t, cmtPk(t, newPk2)))])
		require.NotContains(set, string(mustMarshal(t, cmtPk(t, oldPk1))))
		require.NotContains(set, string(mustMarshal(t, cmtPk(t, oldPk2))))

		// H+2 EndBlock: both rotations drain, no further emits.
		s.ctx = s.ctx.WithBlockHeight(H + 2)
		updates, err = s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Empty(updates)

		// both validators now hold their new cons keys.
		stored1, err := s.stakingKeeper.GetValidator(s.ctx, valAddr1)
		require.NoError(err)
		got1, err := stored1.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(newPk1.Address()).Bytes(), got1)

		stored2, err := s.stakingKeeper.GetValidator(s.ctx, valAddr2)
		require.NoError(err)
		got2, err := stored2.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(newPk2.Address()).Bytes(), got2)
	})
}

// bondedValidatorAtPower stores a bonded validator with the given consensus
// pubkey and seeds every index ApplyAndReturnValidatorSetUpdates reads:
// the operator-keyed validator record, the byConsAddr index, the power
// index, and LastValidatorPower. With LastValidatorPower seeded, the main
// loop will not emit a power-change update unless the test mutates the
// validator's tokens.
func (s *KeeperTestSuite) bondedValidatorAtPower(pk cryptotypes.PubKey, power int64) sdk.ValAddress {
	require := s.Require()
	valAddr := sdk.ValAddress(pk.Address())
	v, err := stakingtypes.NewValidator(valAddr.String(), pk, stakingtypes.Description{Moniker: "v"})
	require.NoError(err)
	v.Status = stakingtypes.Bonded
	v.Tokens = sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction)
	v.DelegatorShares = math.LegacyNewDecFromInt(v.Tokens)
	require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))
	require.NoError(s.stakingKeeper.SetValidatorByConsAddr(s.ctx, v))
	require.NoError(s.stakingKeeper.SetNewValidatorByPowerIndex(s.ctx, v))
	require.NoError(s.stakingKeeper.SetLastValidatorPower(s.ctx, valAddr, power))
	return valAddr
}

// setValidatorPower mutates the bonded validator's tokens to the given
// consensus power and refreshes the by-power index, leaving LastValidatorPower
// untouched so the next EndBlock observes a power change.
func (s *KeeperTestSuite) setValidatorPower(valAddr sdk.ValAddress, power int64) {
	require := s.Require()
	v, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
	require.NoError(err)
	require.NoError(s.stakingKeeper.DeleteValidatorByPowerIndex(s.ctx, v))
	v.Tokens = sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction)
	v.DelegatorShares = math.LegacyNewDecFromInt(v.Tokens)
	require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))
	require.NoError(s.stakingKeeper.SetValidatorByPowerIndex(s.ctx, v))
}

// cmtPk converts a cryptotypes.PubKey to its cmtproto form for direct
// comparison against abci.ValidatorUpdate.PubKey.
func cmtPk(t *testing.T, pk cryptotypes.PubKey) cmtprotocrypto.PublicKey {
	t.Helper()
	out, err := cryptocodec.ToCmtProtoPublicKey(pk)
	if err != nil {
		t.Fatalf("ToCmtProtoPublicKey: %v", err)
	}
	return out
}

// applyCometSet folds a ValidatorUpdate slice into the per-key power map
// CometBFT would derive from it: later updates overwrite earlier ones for
// the same pubkey, and a zero-power update removes the entry. Use this for
// assertions where the slice order is not the load-bearing property.
func applyCometSet(t *testing.T, updates []abci.ValidatorUpdate) map[string]int64 {
	t.Helper()
	set := map[string]int64{}
	for _, u := range updates {
		bz, err := u.PubKey.Marshal()
		if err != nil {
			t.Fatalf("PubKey.Marshal: %v", err)
		}
		if u.Power == 0 {
			delete(set, string(bz))
		} else {
			set[string(bz)] = u.Power
		}
	}
	return set
}

// mustMarshal returns the wire form of a cmt proto pubkey for use as a map key.
func mustMarshal(t *testing.T, pk cmtprotocrypto.PublicKey) []byte {
	t.Helper()
	bz, err := pk.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return bz
}
