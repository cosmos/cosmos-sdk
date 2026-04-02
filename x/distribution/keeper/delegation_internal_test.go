package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
)

func TestNormalizeDelegationStakeWithMargin(t *testing.T) {
	currentStake := math.LegacyOneDec()

	t.Run("keeps current stake for tiny positive delta", func(t *testing.T) {
		stake := currentStake.Add(math.LegacySmallestDec().MulInt64(4))

		normalized := normalizeDelegationStakeWithMargin("delegator", stake, currentStake)
		require.True(t, normalized.Equal(currentStake))
	})

	t.Run("leaves equal-or-lower stake unchanged", func(t *testing.T) {
		equalStake := normalizeDelegationStakeWithMargin("delegator", currentStake, currentStake)
		require.True(t, equalStake.Equal(currentStake))

		lowerStake := currentStake.Sub(math.LegacySmallestDec())
		normalized := normalizeDelegationStakeWithMargin("delegator", lowerStake, currentStake)
		require.True(t, normalized.Equal(lowerStake))
	})

	t.Run("panics when delta exceeds tolerated margin", func(t *testing.T) {
		stake := currentStake.Add(math.LegacySmallestDec().MulInt64(delegationStakeRoundingMarginMultiplier + 1))

		defer func() {
			panicValue := recover()
			require.NotNil(t, panicValue)

			panicMsg, ok := panicValue.(string)
			require.True(t, ok)
			require.Contains(t, panicMsg, "calculated final stake for delegator delegator greater than current stake")
			require.Contains(t, panicMsg, "delta:")
			require.Contains(t, panicMsg, "tolerated delta:")
		}()

		_ = normalizeDelegationStakeWithMargin("delegator", stake, currentStake)
	})
}
