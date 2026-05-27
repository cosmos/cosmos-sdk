package keeper_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmttypes "github.com/cometbft/cometbft/types"
	testrequire "github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type cometValidator struct {
	pk    cryptotypes.PubKey
	power int64
}

func (s *KeeperTestSuite) TestApplyAndReturnValidatorSetUpdates() {
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
		requireCometAppliesValidatorUpdates(t, updates, cometValidator{pk: oldPk, power: 10})

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
		requireCometAppliesValidatorUpdates(t, updates,
			cometValidator{pk: oldPk1, power: 10},
			cometValidator{pk: oldPk2, power: 10},
		)

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

	s.T().Run("validator rotates and changes power at H", func(t *testing.T) {
		s.SetupTest()
		const H = int64(10)
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		s.ctx = s.ctx.WithBlockHeight(H)

		// initial validator set is a oldPk at 10 power
		valAddr := s.bondedValidatorAtPower(oldPk, 10)

		// setup a rotation to newPk
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		// validator also is undergoing a power change from 10 -> 25
		s.setValidatorPower(valAddr, 25)

		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)

		// ensure returned updates apply on top of initialCometVals validator set
		initialCometVals := cometValidator{pk: oldPk, power: 10}
		valSet := requireCometAppliesValidatorUpdates(t, updates, initialCometVals)

		// ensure after applying the updates the old pk is out of the set
		require.False(valSet.HasAddress(oldPk.Address()))

		// ensure after applying the updates the new pk has the updated power
		// including the power change after the rotation
		_, newCmtVal := valSet.GetByAddress(newPk.Address())
		require.NotNil(newCmtVal)
		require.Equal(int64(25), newCmtVal.VotingPower)

		// ensure that after advancing two heights, the staking state is
		// properly updated to reflect the rotation
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

	s.T().Run("validator rotates and is jailed at H", func(t *testing.T) {
		s.SetupTest()
		s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		const H = int64(10)
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		s.ctx = s.ctx.WithBlockHeight(H)

		// initial validator set is a oldPk at 10 power
		valAddr := s.bondedValidatorAtPower(oldPk, 10)

		// setup a rotation to newPk
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		// validator also is being jailed in the same block as the rotation
		require.NoError(s.stakingKeeper.Jail(s.ctx, sdk.ConsAddress(oldPk.Address())))

		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)

		// add a bystander so comet is testing the update batch, not rejecting
		// because the validator set would become empty
		bystander := randomCometValidator()

		// ensure returned updates apply on top of initialCometVals validator set
		valSet := requireCometAppliesValidatorUpdates(t, updates,
			cometValidator{pk: oldPk, power: 10},
			bystander,
		)

		// ensure after applying the updates the old pk is out of the set
		require.False(valSet.HasAddress(oldPk.Address()))

		// ensure after applying the updates the new pk was not added
		// because the validator was jailed before comet applied the rotation
		require.False(valSet.HasAddress(newPk.Address()))

		// ensure that after advancing two heights, the staking state is
		// properly updated to reflect the rotation
		s.ctx = s.ctx.WithBlockHeight(H + stakingtypes.ConsensusUpdateDelay)
		updates, err = s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Empty(updates)

		stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
		require.NoError(err)
		got, err := stored.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(newPk.Address()).Bytes(), got)
	})

	s.T().Run("validator rotates at H and is jailed at H+1", func(t *testing.T) {
		s.SetupTest()
		s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		const H = int64(10)
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()

		s.ctx = s.ctx.WithBlockHeight(H)
		// initial validator set is a oldPk at 10 power
		valAddr := s.bondedValidatorAtPower(oldPk, 10)

		// setup a rotation to newPk
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)

		// ensure returned updates apply on top of initialCometVals validator set
		valSet := requireCometAppliesValidatorUpdates(t, updates, cometValidator{pk: oldPk, power: 10})

		// ensure after applying the updates the old pk is out of the set
		require.False(valSet.HasAddress(oldPk.Address()))

		// ensure after applying the updates the new pk has the original power
		_, newCmtVal := valSet.GetByAddress(newPk.Address())
		require.NotNil(newCmtVal)
		require.Equal(int64(10), newCmtVal.VotingPower)

		// validator is jailed in the block after the rotation update is emitted
		s.ctx = s.ctx.WithBlockHeight(H + 1)
		require.NoError(s.stakingKeeper.Jail(s.ctx, sdk.ConsAddress(oldPk.Address())))
		updates, err = s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)

		// add a bystander so comet is testing the update batch, not rejecting
		// because the validator set would become empty
		bystander := randomCometValidator()
		valSet = requireCometAppliesValidatorUpdates(t, updates,
			cometValidator{pk: newPk, power: 10},
			bystander,
		)

		// ensure after applying the updates the new pk is out of the set
		require.False(valSet.HasAddress(newPk.Address()))

		// ensure that after advancing two heights, the staking state is
		// properly updated to reflect the rotation
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

	s.T().Run("validator rotates and changes power while another validator enters at H", func(t *testing.T) {
		s.SetupTest()
		const H = int64(10)
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		otherPk := ed25519.GenPrivKey().PubKey()
		s.ctx = s.ctx.WithBlockHeight(H)

		// initial validator set is a oldPk at 10 power
		valAddr := s.bondedValidatorAtPower(oldPk, 10)

		// setup a rotation to newPk
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		// validator also is undergoing a power change from 10 -> 15
		s.setValidatorPower(valAddr, 15)

		// another validator is entering the set at 9 power
		s.bondedValidatorAtPowerWithoutLastPower(sdk.ValAddress(otherPk.Address()), otherPk, 9)

		updates, err := s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)

		// ensure returned updates apply on top of initialCometVals validator set
		valSet := requireCometAppliesValidatorUpdates(t, updates, cometValidator{pk: oldPk, power: 10})

		// ensure after applying the updates the old pk is out of the set
		require.False(valSet.HasAddress(oldPk.Address()))

		// ensure after applying the updates the new pk has the updated power
		// including the power change after the rotation
		_, newCmtVal := valSet.GetByAddress(newPk.Address())
		require.NotNil(newCmtVal)
		require.Equal(int64(15), newCmtVal.VotingPower)

		// ensure after applying the updates the other validator is in the set
		_, otherCmtVal := valSet.GetByAddress(otherPk.Address())
		require.NotNil(otherCmtVal)
		require.Equal(int64(9), otherCmtVal.VotingPower)

		// ensure that after advancing two heights, the staking state is
		// properly updated to reflect the rotation
		s.ctx = s.ctx.WithBlockHeight(H + stakingtypes.ConsensusUpdateDelay)
		updates, err = s.stakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
		require.NoError(err)
		require.Empty(updates)

		stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
		require.NoError(err)
		got, err := stored.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(newPk.Address()).Bytes(), got)
	})
}

func (s *KeeperTestSuite) bondedValidatorAtPower(pk cryptotypes.PubKey, power int64) sdk.ValAddress {
	return s.bondedValidatorAtPowerWithOperator(sdk.ValAddress(pk.Address()), pk, power, true)
}

func (s *KeeperTestSuite) bondedValidatorAtPowerWithoutLastPower(
	valAddr sdk.ValAddress,
	pk cryptotypes.PubKey,
	power int64,
) sdk.ValAddress {
	return s.bondedValidatorAtPowerWithOperator(valAddr, pk, power, false)
}

func (s *KeeperTestSuite) bondedValidatorAtPowerWithOperator(
	valAddr sdk.ValAddress,
	pk cryptotypes.PubKey,
	power int64,
	seedLastPower bool,
) sdk.ValAddress {
	require := s.Require()
	v, err := stakingtypes.NewValidator(valAddr.String(), pk, stakingtypes.Description{Moniker: "v"})
	require.NoError(err)
	v.Status = stakingtypes.Bonded
	v.Tokens = sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction)
	v.DelegatorShares = math.LegacyNewDecFromInt(v.Tokens)
	require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))
	require.NoError(s.stakingKeeper.SetValidatorByConsAddr(s.ctx, v))
	require.NoError(s.stakingKeeper.SetNewValidatorByPowerIndex(s.ctx, v))
	if seedLastPower {
		require.NoError(s.stakingKeeper.SetLastValidatorPower(s.ctx, valAddr, power))
	}
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

func randomCometValidator() cometValidator {
	return cometValidator{
		pk:    ed25519.GenPrivKey().PubKey(),
		power: 1,
	}
}

func requireCometAppliesValidatorUpdates(
	t *testing.T,
	updates []abci.ValidatorUpdate,
	initialValidators ...cometValidator,
) *cmttypes.ValidatorSet {
	t.Helper()

	initialUpdates := make([]abci.ValidatorUpdate, 0, len(initialValidators))
	for _, validator := range initialValidators {
		initialUpdates = append(initialUpdates, abci.ValidatorUpdate{
			PubKey: cmtPk(t, validator.pk),
			Power:  validator.power,
		})
	}

	initialCmtVals, err := cmttypes.PB2TM.ValidatorUpdates(initialUpdates)
	testrequire.NoError(t, err)
	valSet := cmttypes.NewValidatorSet(initialCmtVals)

	cmtUpdates, err := cmttypes.PB2TM.ValidatorUpdates(updates)
	testrequire.NoError(t, err)
	testrequire.NoError(t, valSet.UpdateWithChangeSet(cmtUpdates))
	return valSet
}
