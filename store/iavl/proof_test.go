package iavl

import (
	"bytes"
	"testing"

	ics23 "github.com/confio/ics23/go"
)

func TestConvertExistence(t *testing.T) {
	proof, err := GenerateIavlResult(200, Middle)
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

func TestCreateMembership(t *testing.T) {
	cases := map[string]struct {
		size int
		loc  Where
	}{
		"small left":   {size: 100, loc: Left},
		"small middle": {size: 100, loc: Middle},
		"small right":  {size: 100, loc: Right},
		"big left":     {size: 5431, loc: Left},
		"big middle":   {size: 5431, loc: Middle},
		"big right":    {size: 5431, loc: Right},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tree, allkeys, err := BuildTree(tc.size)
			if err != nil {
				t.Fatalf("Creating tree: %+v", err)
			}
			key := GetKey(allkeys, tc.loc)
			_, val := tree.Get(key)
			proof, err := CreateMembershipProof(tree, key)
			if err != nil {
				t.Fatalf("Creating Proof: %+v", err)
			}

			root := tree.Hash()
			valid := ics23.VerifyMembership(ics23.IavlSpec, root, proof, key, val)
			if !valid {
				t.Fatalf("Membership Proof Invalid")
			}
		})
	}
}

func TestCreateNonMembership(t *testing.T) {
	cases := map[string]struct {
		size int
		loc  Where
	}{
		"small left":   {size: 100, loc: Left},
		"small middle": {size: 100, loc: Middle},
		"small right":  {size: 100, loc: Right},
		"big left":     {size: 5431, loc: Left},
		"big middle":   {size: 5431, loc: Middle},
		"big right":    {size: 5431, loc: Right},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tree, allkeys, err := BuildTree(tc.size)
			if err != nil {
				t.Fatalf("Creating tree: %+v", err)
			}
			key := GetNonKey(allkeys, tc.loc)

			proof, err := CreateNonMembershipProof(tree, key)
			if err != nil {
				t.Fatalf("Creating Proof: %+v", err)
			}

			root := tree.Hash()
			valid := ics23.VerifyNonMembership(ics23.IavlSpec, root, proof, key)
			if !valid {
				t.Fatalf("Non Membership Proof Invalid")
			}
		})
	}
}
