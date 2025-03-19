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
			require.True(t, budget.BudgetPerTranche.Equal(tc.expBudget.BudgetPerTranche),
				fmt.Sprintf("expected BudgetPerTranche %v, got %v", tc.expBudget.BudgetPerTranche, budget.BudgetPerTranche))
			require.Equal(t, tc.expBudget.RecipientAddress, budget.RecipientAddress)
			require.True(t, budget.LastClaimedAt.Equal(tc.expBudget.LastClaimedAt),
				fmt.Sprintf("expected LastClaimedAt %v, got %v", tc.expBudget.LastClaimedAt, budget.LastClaimedAt))
			require.Equal(t, tc.expBudget.TranchesLeft, budget.TranchesLeft)
			require.Equal(t, tc.expBudget.Period, budget.Period)
		})
	}
}

// TestValidateAmount tests the validateAmount function.
func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount sdk.Coins
		expErr bool
		errMsg string
	}{
		{
			name:   "nil amount",
			amount: nil,
			expErr: true,
			errMsg: "amount cannot be nil",
		},
		{
			name: "negative coin amount",
			amount: sdk.Coins{
				{
					Denom:  "stake",
					Amount: math.NewInt(-100),
				},
			},
			expErr: true,
			errMsg: "-100",
		},
		{
			name:   "valid single coin",
			amount: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			expErr: false,
		},
		{
			name: "multiple valid coins",
			amount: sdk.NewCoins(
				sdk.NewCoin("stake", math.NewInt(100)),
				sdk.NewCoin("token", math.NewInt(200)),
			),
			expErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAmount(tc.amount)
			if tc.expErr {
				require.Error(t, err, "expected an error but got none")
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateContinuousFund(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	// Create a context with the current block time.
	ctx := sdk.Context{}.WithBlockTime(now)

	tests := []struct {
		name   string
		msg    types.MsgCreateContinuousFund
		expErr bool
		errMsg string
	}{
		{
			name: "zero percentage",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyZeroDec(),
				Expiry:     &future,
			},
			expErr: true,
			errMsg: "percentage cannot be zero or empty",
		},
		{
			name: "negative percentage",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyNewDecFromInt(math.NewInt(-1)),
				Expiry:     &future,
			},
			expErr: true,
			errMsg: "percentage cannot be negative",
		},
		{
			name: "percentage equal to one",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyOneDec(),
				Expiry:     &future,
			},
			expErr: true,
			errMsg: "percentage cannot be greater than or equal to one",
		},
		{
			name: "valid percentage with nil expiry",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyNewDecWithPrec(5, 1), // 0.5
				Expiry:     nil,
			},
			expErr: false,
		},
		{
			name: "valid percentage with future expiry",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyNewDecWithPrec(5, 1), // 0.5
				Expiry:     &future,
			},
			expErr: false,
		},
		{
			name: "expiry in past",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyNewDecWithPrec(5, 1), // 0.5
				Expiry:     &past,
			},
			expErr: true,
			errMsg: "expiry time cannot be less than the current block time",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateContinuousFund(ctx, tc.msg)
			if tc.expErr {
				require.Error(t, err, "expected an error but got none")
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
