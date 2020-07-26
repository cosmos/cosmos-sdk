package simulation_test

import (
	"fmt"
	"testing"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/KiraCore/cosmos-sdk/simapp"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/x/auth/simulation"
	"github.com/KiraCore/cosmos-sdk/x/auth/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
)

func TestDecodeStore(t *testing.T) {
	app := simapp.Setup(false)
	cdc, _ := simapp.MakeCodecs()
	acc := types.NewBaseAccountWithAddress(delAddr1)
	dec := simulation.NewDecodeStore(app.AccountKeeper)

	accBz, err := app.AccountKeeper.MarshalAccount(acc)
	require.NoError(t, err)

	globalAccNumber := gogotypes.UInt64Value{Value: 10}

	kvPairs := tmkv.Pairs{
		tmkv.Pair{
			Key:   types.AddressStoreKey(delAddr1),
			Value: accBz,
		},
		tmkv.Pair{
			Key:   types.GlobalAccountNumberKey,
			Value: cdc.MustMarshalBinaryBare(&globalAccNumber),
		},
		tmkv.Pair{
			Key:   []byte{0x99},
			Value: []byte{0x99},
		},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Account", fmt.Sprintf("%v\n%v", acc, acc)},
		{"GlobalAccNumber", fmt.Sprintf("GlobalAccNumberA: %d\nGlobalAccNumberB: %d", globalAccNumber, globalAccNumber)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
