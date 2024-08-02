package simapp

import (
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateWeightedValidators(t *testing.T) {
	doHash := func(b ...byte) []byte {
		hash := sha256.Sum256(b)
		return hash[:20]
	}
	val1, val2, val3 := WeightedValidator{
		power: 100,
		addr:  doHash(1),
	}, WeightedValidator{
		power: 100,
		addr:  doHash(2),
	}, WeightedValidator{
		power: 50,
		addr:  doHash(3),
	}
	specs := map[string]struct {
		updates []appmodulev2.ValidatorUpdate
		exp     WeightedValidators
	}{
		"no updates": {
			updates: []appmodulev2.ValidatorUpdate{},
			exp:     WeightedValidators{val1, val2, val3},
		},
		"add one": {
			updates: []appmodulev2.ValidatorUpdate{{PubKey: []byte{4}, Power: 200}},
			exp:     WeightedValidators{WeightedValidator{addr: doHash(4), power: 200}, val1, val2, val3},
		},
		"remove one": {
			updates: []appmodulev2.ValidatorUpdate{{PubKey: []byte{2}, Power: 0}},
			exp:     WeightedValidators{val1, val3},
		},
		"update one": {
			updates: []appmodulev2.ValidatorUpdate{{PubKey: []byte{2}, Power: 20}},
			exp:     WeightedValidators{val1, val3, WeightedValidator{addr: doHash(2), power: 20}},
		},
		"multiple operations": {
			updates: []appmodulev2.ValidatorUpdate{{PubKey: []byte{1}, Power: 20}, {PubKey: []byte{3}, Power: 0}},
			exp:     WeightedValidators{val2, WeightedValidator{addr: doHash(1), power: 20}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			src := NewValSet(val1, val2, val3)
			gotValset := src.Update(spec.updates)
			assert.Equal(t, spec.exp, gotValset)
		})
	}
}
