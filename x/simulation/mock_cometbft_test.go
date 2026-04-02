package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/stretchr/testify/require"
)

func TestUpdateValidatorsPreventsEmptySet(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	params := RandomParams(r)

	existing := newValidatorUpdate(1, 10)
	existingKey := validatorKey(existing)

	current := map[string]mockValidator{
		existingKey: {
			val:           existing,
			livenessState: 0,
		},
	}

	var events []string
	next := updateValidators(
		t,
		r,
		params,
		current,
		[]abci.ValidatorUpdate{{PubKey: existing.PubKey, Power: 0}},
		func(route, op, result string) {
			events = append(events, fmt.Sprintf("%s/%s/%s", route, op, result))
		},
	)

	require.Len(t, next, 1)
	_, kept := next[existingKey]
	require.True(t, kept, "last validator should be preserved to avoid empty validator set")
	require.Contains(t, events, "end_block/validator_updates/prevented-empty-set")
}

func TestUpdateValidatorsDeleteThenAddKeepsNonEmptyWithoutFallback(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	params := RandomParams(r)

	existing := newValidatorUpdate(1, 10)
	replacement := newValidatorUpdate(2, 20)
	existingKey := validatorKey(existing)
	replacementKey := validatorKey(replacement)

	current := map[string]mockValidator{
		existingKey: {
			val:           existing,
			livenessState: 0,
		},
	}

	var events []string
	next := updateValidators(
		t,
		r,
		params,
		current,
		[]abci.ValidatorUpdate{
			{PubKey: existing.PubKey, Power: 0},
			replacement,
		},
		func(route, op, result string) {
			events = append(events, fmt.Sprintf("%s/%s/%s", route, op, result))
		},
	)

	require.Len(t, next, 1)
	_, oldExists := next[existingKey]
	require.False(t, oldExists)
	_, newExists := next[replacementKey]
	require.True(t, newExists)
	require.NotContains(t, events, "end_block/validator_updates/prevented-empty-set")
}

func newValidatorUpdate(byteVal byte, power int64) abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		PubKey: cmtproto.PublicKey{
			Sum: &cmtproto.PublicKey_Ed25519{
				Ed25519: bytesOf(byteVal, 32),
			},
		},
		Power: power,
	}
}

func bytesOf(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}

func validatorKey(v abci.ValidatorUpdate) string {
	return fmt.Sprintf("%X", v.PubKey.GetEd25519())
}
