package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// createInactiveProposals creates n inactive proposals (in deposit period) for testing
func createInactiveProposals(t *testing.T, ctx sdk.Context, k *keeper.Keeper, n int) {
	for i := 0; i < n; i++ {
		tp := TestProposal
		// Create a test address for the proposer
		proposer := sdk.AccAddress([]byte("proposer" + string(rune(i))))
		_, err := k.SubmitProposal(ctx, tp, "", "title", "summary", proposer)
		require.NoError(t, err)
		// Proposals are in deposit period by default after submission
	}
}

func TestGetMinInitialDeposit(t *testing.T) {
	var (
		minInitialDepositFloor   = v1.GetDefaultMinInitialDepositFloor()
		minInitialDepositFloorX2 = minInitialDepositFloor.MulInt(math.NewInt(2))
		updatePeriod             = v1.DefaultMinInitialDepositUpdatePeriod
		N                        = int(v1.DefaultTargetProposalsInDepositPeriod)

		minInitialDepositTimeFromTicks = func(ticks int) *time.Time {
			t := time.Now().Add(-updatePeriod*time.Duration(ticks) - time.Minute)
			return &t
		}
	)
	tests := []struct {
		name                      string
		setup                     func(sdk.Context, *keeper.Keeper)
		expectedMinInitialDeposit string
	}{
		{
			name:                      "initial case no setup : expectedMinInitialDeposit=minInitialDepositFloor",
			expectedMinInitialDeposit: minInitialDepositFloor.String(),
		},
		{
			name: "n=N-1 lastMinInitialDeposit=minInitialDepositFloor ticksPassed=0 : expectedMinInitialDeposit=minInitialDepositFloor",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N-1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloor,
					Time:  minInitialDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, false)
			},
			expectedMinInitialDeposit: minInitialDepositFloor.String(),
		},
		{
			name: "n=N lastMinInitialDeposit=minInitialDepositFloor ticksPassed=0 : expectedMinInitialDeposit>minInitialDepositFloor",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloor,
					Time:  minInitialDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, false)
			},
			expectedMinInitialDeposit: "101000stake",
		},
		{
			name: "n=N+1 lastMinInitialDeposit=minInitialDepositFloor ticksPassed=0 : expectedMinInitialDeposit>minInitialDepositFloor",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N+1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloor,
					Time:  minInitialDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, false)
			},
			expectedMinInitialDeposit: "101000stake",
		},
		{
			name: "n=N+1 lastMinInitialDeposit=otherCoins ticksPassed=0 : expectedMinInitialDeposit>minInitialDepositFloor",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N+1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: sdk.NewCoins(
						sdk.NewInt64Coin("xxx", 1_000_000_000),
					),
					Time: minInitialDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, false)
			},
			expectedMinInitialDeposit: "101000stake",
		},
		{
			name: "n=N-1 lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=0 : expectedMinInitialDeposit=minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N-1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, false)
			},
			expectedMinInitialDeposit: minInitialDepositFloorX2.String(),
		},
		{
			name: "n=N lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=0 : expectedMinInitialDeposit>minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, false)
			},
			expectedMinInitialDeposit: "202000stake",
		},
		{
			name: "n=N+1 lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=0 : expectedMinInitialDeposit>minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N+1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, false)
			},
			expectedMinInitialDeposit: "202000stake",
		},
		{
			name: "n=N+1 lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=0 (try time-based update) : expectedMinInitialDeposit=minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N+1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(0),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, true)
			},
			expectedMinInitialDeposit: minInitialDepositFloorX2.String(),
		},
		{
			name: "n=N-1 lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=1 : expectedMinInitialDeposit<minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N-1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(1),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, true)
			},
			expectedMinInitialDeposit: "199000stake",
		},
		{
			name: "n=N lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=1 : expectedMinInitialDeposit=minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(1),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, true)
			},
			expectedMinInitialDeposit: minInitialDepositFloorX2.String(),
		},
		{
			name: "n=N+1 lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=1 : expectedMinInitialDeposit=minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N+1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(1),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx, true)
			},
			expectedMinInitialDeposit: minInitialDepositFloorX2.String(),
		},
		{
			name: "n=N-1 lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=2 : expectedMinInitialDeposit<minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N-1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(2),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx.WithBlockTime(*minInitialDepositTimeFromTicks(1)), true)
				k.UpdateMinInitialDeposit(ctx, true)
			},
			expectedMinInitialDeposit: "198005stake",
		},
		{
			name: "n=N lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=2 : expectedMinInitialDeposit=minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(2),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx.WithBlockTime(*minInitialDepositTimeFromTicks(1)), true)
				k.UpdateMinInitialDeposit(ctx, true)
			},
			expectedMinInitialDeposit: minInitialDepositFloorX2.String(),
		},
		{
			name: "n=N+1 lastMinInitialDeposit=minInitialDepositFloor*2 ticksPassed=2 : expectedMinInitialDeposit=minInitialDepositFloor*2",
			setup: func(ctx sdk.Context, k *keeper.Keeper) {
				createInactiveProposals(t, ctx, k, N+1)
				err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
					Value: minInitialDepositFloorX2,
					Time:  minInitialDepositTimeFromTicks(2),
				})
				require.NoError(t, err)
				k.UpdateMinInitialDeposit(ctx.WithBlockTime(*minInitialDepositTimeFromTicks(1)), true)
				k.UpdateMinInitialDeposit(ctx, true)
			},
			expectedMinInitialDeposit: minInitialDepositFloorX2.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, _, _, _, _, _, ctx := setupGovKeeper(t)
			if tt.setup != nil {
				tt.setup(ctx, k)
			}

			minInitialDeposit := k.GetMinInitialDeposit(ctx)

			assert.Equal(t, tt.expectedMinInitialDeposit, minInitialDeposit.String())
		})
	}
}
