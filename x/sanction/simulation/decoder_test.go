package simulation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction/simulation"
)

func TestDecodeStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Codec
	dec := simulation.NewDecodeStore(cdc)

	tests := []struct {
		name     string
		kvA      kv.Pair
		kvB      kv.Pair
		exp      string
		expPanic string
	}{
		{
			name: "params",
			kvA:  kv.Pair{Key: keeper.CreateParamKey("pnamea"), Value: []byte("valuea")},
			kvB:  kv.Pair{Key: keeper.CreateParamKey("pnameb"), Value: []byte("valueb")},
			exp:  "valuea\nvalueb",
		},
		{
			name: "sanction",
			kvA:  kv.Pair{Key: keeper.CreateSanctionedAddrKey(sdk.AccAddress("addra")), Value: []byte{50}},
			kvB:  kv.Pair{Key: keeper.CreateSanctionedAddrKey(sdk.AccAddress("addrb")), Value: []byte{51}},
			exp:  "[50]\n[51]",
		},
		{
			name: "temp",
			kvA:  kv.Pair{Key: keeper.CreateTemporaryKey(sdk.AccAddress("addra"), 1), Value: []byte{52}},
			kvB:  kv.Pair{Key: keeper.CreateTemporaryKey(sdk.AccAddress("addrb"), 2), Value: []byte{53}},
			exp:  "[52]\n[53]",
		},
		{
			name: "index",
			kvA:  kv.Pair{Key: keeper.CreateProposalTempIndexKey(1, sdk.AccAddress("addra")), Value: []byte{54}},
			kvB:  kv.Pair{Key: keeper.CreateProposalTempIndexKey(1, sdk.AccAddress("addrb")), Value: []byte{55}},
			exp:  "[54]\n[55]",
		},
		{
			name:     "unknown",
			kvA:      kv.Pair{Key: []byte{0x9a}, Value: []byte("valuea")},
			kvB:      kv.Pair{Key: []byte{0x9c}, Value: []byte("valueb")},
			expPanic: "invalid sanction key 9A",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = dec(tc.kvA, tc.kvB)
			}
			if len(tc.expPanic) > 0 {
				assert.PanicsWithValue(t, tc.expPanic, testFunc, "running decoder")
			} else {
				if assert.NotPanics(t, testFunc, "running decoder") {
					assert.Equal(t, tc.exp, actual, "decoder result")
				}
			}
		})
	}
}
