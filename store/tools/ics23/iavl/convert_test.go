package iavlproofs

import (
	"bytes"
	"testing"

	tmproofs "github.com/cosmos/cosmos-sdk/store/internal/proofs"
	"github.com/cosmos/cosmos-sdk/store/tools/ics23/iavl/helpers"
)

func TestConvertExistence(t *testing.T) {
	proof, err := helpers.GenerateIavlResult(200, tmproofs.Middle)
	if err != nil {
		t.Fatal(err)
	}

	converted, err := convertExistenceProof(proof.Proof, proof.Key, proof.Value)
	if err != nil {
		t.Fatal(err)
	}

	calc, err := converted.Calculate()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(calc, proof.RootHash) {
		t.Errorf("Calculated: %X\nExpected:   %X", calc, proof.RootHash)
	}
}
