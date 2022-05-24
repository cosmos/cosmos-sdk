package simulation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/cosmos-sdk/crypto/keys/ed25519"
	"github.com/Stride-Labs/cosmos-sdk/simapp"
	sdk "github.com/Stride-Labs/cosmos-sdk/types"
	"github.com/Stride-Labs/cosmos-sdk/types/kv"
	"github.com/Stride-Labs/cosmos-sdk/x/evidence/simulation"
	"github.com/Stride-Labs/cosmos-sdk/x/evidence/types"
)

func TestDecodeStore(t *testing.T) {
	app := simapp.Setup(false)
	dec := simulation.NewDecodeStore(app.EvidenceKeeper)

	delPk1 := ed25519.GenPrivKey().PubKey()

	ev := &types.Equivocation{
		Height:           10,
		Time:             time.Now().UTC(),
		Power:            1000,
		ConsensusAddress: sdk.ConsAddress(delPk1.Address()).String(),
	}

	evBz, err := app.EvidenceKeeper.MarshalEvidence(ev)
	require.NoError(t, err)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{
				Key:   types.KeyPrefixEvidence,
				Value: evBz,
			},
			{
				Key:   []byte{0x99},
				Value: []byte{0x99},
			},
		},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Evidence", fmt.Sprintf("%v\n%v", ev, ev)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
