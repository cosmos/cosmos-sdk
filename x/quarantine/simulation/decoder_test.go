package simulation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/simulation"
)

func TestDecodeStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Codec
	dec := simulation.NewDecodeStore(cdc)

	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	marshal := func(o codec.ProtoMarshaler, name string) []byte {
		rv, err := cdc.Marshal(o)
		require.NoError(t, err, "cdc.Marshal(%s)", name)
		return rv
	}

	addr0 := sdk.AccAddress("addr0_______________")
	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")
	addr3 := sdk.AccAddress("addr3_______________")

	autoRespA := quarantine.AUTO_RESPONSE_ACCEPT
	autoRespB := quarantine.AUTO_RESPONSE_DECLINE
	autoRespABz := []byte{quarantine.ToAutoB(autoRespA)}
	autoRespBBz := []byte{quarantine.ToAutoB(autoRespB)}

	recordA := quarantine.NewQuarantineRecord([]string{addr1.String()}, cz("5bananas"), false)
	recordB := quarantine.NewQuarantineRecord([]string{addr3.String()}, cz("8sunflowers"), true)
	recordABz := marshal(recordA, "recordA")
	recordBBz := marshal(recordB, "recordB")

	recordIndexA := &quarantine.QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{[]byte("0123"), []byte("6789")}}
	recordIndexB := &quarantine.QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{[]byte("abcd"), []byte("wxyz")}}
	recordIndexABz := marshal(recordIndexA, "recordIndexA")
	recordIndexBBz := marshal(recordIndexB, "recordIndexB")

	tests := []struct {
		name     string
		kvA      kv.Pair
		kvB      kv.Pair
		exp      string
		expPanic string
	}{
		{
			name: "OptIn",
			kvA:  kv.Pair{Key: quarantine.CreateOptInKey(addr0), Value: []byte{0x00}},
			kvB:  kv.Pair{Key: quarantine.CreateOptInKey(addr1), Value: []byte{0x01}},
			exp:  "[0]\n[1]",
		},
		{
			name: "AutoResponse",
			kvA:  kv.Pair{Key: quarantine.CreateAutoResponseKey(addr0, addr1), Value: autoRespABz},
			kvB:  kv.Pair{Key: quarantine.CreateAutoResponseKey(addr2, addr3), Value: autoRespBBz},
			exp:  "AUTO_RESPONSE_ACCEPT\nAUTO_RESPONSE_DECLINE",
		},
		{
			name: "Record",
			kvA:  kv.Pair{Key: quarantine.CreateRecordKey(addr0, addr1), Value: recordABz},
			kvB:  kv.Pair{Key: quarantine.CreateRecordKey(addr2, addr3), Value: recordBBz},
			exp:  "{[61646472315F5F5F5F5F5F5F5F5F5F5F5F5F5F5F] [] 5bananas false}\n{[61646472335F5F5F5F5F5F5F5F5F5F5F5F5F5F5F] [] 8sunflowers true}",
		},
		{
			name: "RecordIndex",
			kvA:  kv.Pair{Key: quarantine.CreateRecordIndexKey(addr0, addr1), Value: recordIndexABz},
			kvB:  kv.Pair{Key: quarantine.CreateRecordIndexKey(addr1, addr2), Value: recordIndexBBz},
			exp:  "{[[48 49 50 51] [54 55 56 57]]}\n{[[97 98 99 100] [119 120 121 122]]}",
		},
		{
			name:     "unknown",
			kvA:      kv.Pair{Key: []byte{0x9a}, Value: []byte{0x9b}},
			kvB:      kv.Pair{Key: []byte{0x9c}, Value: []byte{0x9d}},
			expPanic: "invalid quarantine key 9A",
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
