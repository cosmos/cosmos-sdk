package keeper

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	"github.com/stretchr/testify/require"
)

// TestValidateAndUpdateBudgetProposal tests the validateAndUpdateBudgetProposal function.
func TestValidateAndUpdateBudgetProposal(t *testing.T) {
	// Set up some reference times.
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	// Create a base context with a block time set to 'now'.
	baseCtx := sdk.Context{}.WithBlockTime(now)

	tests := []struct {
		name      string
		ctx       sdk.Context
		bp        types.MsgSubmitBudgetProposal
		expBudget types.Budget
		expErr    bool
		errMsg    string // expected error substring
	}{
		{
			name: "zero budget per tranche",
			ctx:  baseCtx,
			bp: types.MsgSubmitBudgetProposal{
				BudgetPerTranche: sdk.NewCoin("stake", math.NewInt(0)),
				StartTime:        &future,
				Tranches:         1,
				Period:           1,
				RecipientAddress: "cosmos1xxxxxx",
			},
			expErr: true,
			errMsg: "budget per tranche cannot be zero",
		},
		{
			name: "invalid coin amount (negative)",
			ctx:  baseCtx,
			bp: types.MsgSubmitBudgetProposal{
				// Assuming validateAmount will reject negative coins.
				BudgetPerTranche: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(-100)},
				StartTime:        &future,
				Tranches:         1,
				Period:           1,
				RecipientAddress: "cosmos1xxxxxx",
			},
			expErr: true,
			errMsg: "invalid budget proposal",
		},
		{
			name: "start time in past",
			ctx:  baseCtx,
			bp: types.MsgSubmitBudgetProposal{
				BudgetPerTranche: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
				StartTime:        &past,
				Tranches:         1,
				Period:           1,
				RecipientAddress: "cosmos1xxxxxx",
			},
			expErr: true,
			errMsg: "start time cannot be less than the current block time",
		},
		{
			name: "zero tranches",
			ctx:  baseCtx,
			bp: types.MsgSubmitBudgetProposal{
				BudgetPerTranche: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
				StartTime:        &future,
				Tranches:         0,
				Period:           1,
				RecipientAddress: "cosmos1xxxxxx",
			},
			expErr: true,
			errMsg: "tranches must be greater than zero",
		},
		{
			name: "zero period",
			ctx:  baseCtx,
			bp: types.MsgSubmitBudgetProposal{
				BudgetPerTranche: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
				StartTime:        &future,
				Tranches:         1,
				Period:           0,
				RecipientAddress: "cosmos1xxxxxx",
			},
			expErr: true,
			errMsg: "period length should be greater than zero",
		},
		{
			name: "valid proposal with explicit start time",
			ctx:  sdk.Context{}.WithBlockTime(now),
			bp: types.MsgSubmitBudgetProposal{
				BudgetPerTranche: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
				StartTime:        &future,
				Tranches:         5,
				Period:           10,
				RecipientAddress: "cosmos1recipient",
			},
			expErr: false,
			expBudget: types.Budget{
				RecipientAddress: "cosmos1recipient",
				BudgetPerTranche: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
				LastClaimedAt:    future,
				TranchesLeft:     5,
				Period:           10,
			},
		},
		{
			name: "valid proposal with nil start time",
			ctx:  sdk.Context{}.WithBlockTime(now),
			bp: types.MsgSubmitBudgetProposal{
				BudgetPerTranche: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
				StartTime:        nil, // should be set to current block time
				Tranches:         5,
				Period:           10,
				RecipientAddress: "cosmos1recipient",
			},
			expErr: false,
			expBudget: types.Budget{
				RecipientAddress: "cosmos1recipient",
				BudgetPerTranche: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
				LastClaimedAt:    now, // function should set StartTime to ctx.BlockTime()
				TranchesLeft:     5,
				Period:           10,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Execute the function under test.
			budget, err := validateAndUpdateBudgetProposal(tc.ctx, tc.bp)

			if tc.expErr {
				require.Error(t, err, "expected an error but got none")
				require.Contains(t, err.Error(), tc.errMsg)
				return
			}

			require.NoError(t, err)

			// Compare the returned budget fields using require for assertions.
			require.True(t, budget.BudgetPerTranche.IsEqual(tc.expBudget.BudgetPerTranche),
				fmt.Sprintf("expected BudgetPerTranche %v, got %v", tc.expBudget.BudgetPerTranche, budget.BudgetPerTranche))
			require.Equal(t, tc.expBudget.RecipientAddress, budget.RecipientAddress)
			require.True(t, budget.LastClaimedAt.Equal(tc.expBudget.LastClaimedAt),
				fmt.Sprintf("expected LastClaimedAt %v, got %v", tc.expBudget.LastClaimedAt, budget.LastClaimedAt))
			require.Equal(t, tc.expBudget.TranchesLeft, budget.TranchesLeft)
			require.Equal(t, tc.expBudget.Period, budget.Period)
		})
	}
}
