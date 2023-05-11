package feegrant_test

import (
	"testing"
	"time"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

func TestFilteredFeeValidAllow(t *testing.T) {
	app := simapp.Setup(false)

	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 488))
	bigAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1000))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))

	basicAllowance, _ := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{
		SpendLimit: bigAtom,
	})

	cases := map[string]struct {
		allowance *feegrant.AllowedMsgAllowance
		// all other checks are ignored if valid=false
		fee       sdk.Coins
		blockTime time.Time
		valid     bool
		accept    bool
		remove    bool
		remains   sdk.Coins
	}{
		"internal fee is updated": {
			allowance: &feegrant.AllowedMsgAllowance{
				Allowance:       basicAllowance,
				AllowedMessages: []string{"/cosmos.bank.v1beta1.MsgSend"},
			},
			fee:     smallAtom,
			accept:  true,
			remove:  false,
			remains: leftAtom,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			err := tc.allowance.ValidateBasic()
			require.NoError(t, err)

			ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(tc.blockTime)

			// now try to deduct
			removed, err := tc.allowance.Accept(ctx, tc.fee, []sdk.Msg{
				&bank.MsgSend{
					FromAddress: "gm",
					ToAddress:   "gn",
					Amount:      tc.fee,
				},
			})
			if !tc.accept {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.remove, removed)
			if !removed {
				var basicAllowanceLeft feegrant.BasicAllowance
				app.AppCodec().Unmarshal(tc.allowance.Allowance.Value, &basicAllowanceLeft)

				assert.Equal(t, tc.remains, basicAllowanceLeft.SpendLimit)
			}
		})
	}
}
