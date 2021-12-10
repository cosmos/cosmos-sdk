package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/x/group/simulation"
)

func TestDecodeStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Codec
	dec := simulation.NewDecodeStore(cdc)

	g := group.GroupInfo{}
	groupBz, err := cdc.Marshal(&g)
	require.NoError(t, err)
	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: append([]byte{keeper.GroupTablePrefix}, orm.PrimaryKey(&g)...), Value: groupBz},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectErr   bool
		expectedLog string
	}{
		{"Group", false, fmt.Sprintf("%v\n%v", g, g)},
		{"other", true, ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectErr {
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			} else {
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
