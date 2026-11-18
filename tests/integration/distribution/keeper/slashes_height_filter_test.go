package keeper_test

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// TestGRPCValidatorSlashesFiltersByHeight checks that StartingHeight and EndingHeight
// select slash events by the block height they were recorded at (the height in the
// store key), not by the validator period stored in the event.
func TestGRPCValidatorSlashesFiltersByHeight(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	queryClient := types.NewQueryClient(f.app.QueryHelper())

	// slashes at heights 1000, 2000 and 3000, in validator periods 2, 3 and 4
	slashes := []types.ValidatorSlashEvent{
		types.NewValidatorSlashEvent(2, sdkmath.LegacyNewDecWithPrec(1, 2)),
		types.NewValidatorSlashEvent(3, sdkmath.LegacyNewDecWithPrec(2, 2)),
		types.NewValidatorSlashEvent(4, sdkmath.LegacyNewDecWithPrec(3, 2)),
	}
	for i, slash := range slashes {
		height := uint64(i+1) * 1000
		assert.NilError(t, f.distrKeeper.SetValidatorSlashEvent(f.sdkCtx, f.valAddr, height, slash.ValidatorPeriod, slash))
	}

	testCases := []struct {
		name       string
		start, end uint64
		pagination *query.PageRequest
		exp        []types.ValidatorSlashEvent
	}{
		{name: "range around one height", start: 900, end: 1100, exp: slashes[:1]},
		{name: "bounds are inclusive", start: 1000, end: 3000, exp: slashes},
		{name: "single height", start: 2000, end: 2000, exp: slashes[1:2]},
		{name: "between heights", start: 1001, end: 2999, exp: slashes[1:2]},
		{name: "range covering only the periods", start: 1, end: 10, exp: nil},
		{name: "full range", start: 0, end: math.MaxUint64, exp: slashes},
		{name: "limit", start: 1000, end: 3000, pagination: &query.PageRequest{Limit: 2}, exp: slashes[:2]},
		{name: "offset", start: 1000, end: 3000, pagination: &query.PageRequest{Offset: 1, Limit: 1}, exp: slashes[1:2]},
		{name: "limit with range", start: 1500, end: 3000, pagination: &query.PageRequest{Limit: 1}, exp: slashes[1:2]},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := queryClient.ValidatorSlashes(f.sdkCtx, &types.QueryValidatorSlashesRequest{
				ValidatorAddress: f.valAddr.String(),
				StartingHeight:   tc.start,
				EndingHeight:     tc.end,
				Pagination:       tc.pagination,
			})
			assert.NilError(t, err)
			assert.DeepEqual(t, tc.exp, res.Slashes)
		})
	}
}
