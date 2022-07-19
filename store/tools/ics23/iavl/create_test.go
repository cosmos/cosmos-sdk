package iavlproofs

import (
	"testing"

	ics23 "github.com/confio/ics23/go"

	tmproofs "github.com/cosmos/cosmos-sdk/store/internal/proofs"
	"github.com/cosmos/cosmos-sdk/store/tools/ics23/iavl/helpers"
)

func TestCreateMembership(t *testing.T) {
	cases := map[string]struct {
		size int
		loc  tmproofs.Where
	}{
		"small left":   {size: 100, loc: tmproofs.Left},
		"small middle": {size: 100, loc: tmproofs.Middle},
		"small right":  {size: 100, loc: tmproofs.Right},
		"big left":     {size: 5431, loc: tmproofs.Left},
		"big middle":   {size: 5431, loc: tmproofs.Middle},
		"big right":    {size: 5431, loc: tmproofs.Right},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tree, allkeys, err := helpers.BuildTree(tc.size)
			if err != nil {
				t.Fatalf("Creating tree: %+v", err)
			}
			key := helpers.GetKey(allkeys, tc.loc)
			_, val := tree.Get(key)
			proof, err := CreateMembershipProof(tree, key)
			if err != nil {
				t.Fatalf("Creating Proof: %+v", err)
			}

			root := tree.WorkingHash()
			valid := ics23.VerifyMembership(IavlSpec, root, proof, key, val)
			if !valid {
				t.Fatalf("Membership Proof Invalid")
			}
		})
	}
}

func TestCreateNonMembership(t *testing.T) {
	cases := map[string]struct {
		size int
		loc  tmproofs.Where
	}{
		"small left":   {size: 100, loc: tmproofs.Left},
		"small middle": {size: 100, loc: tmproofs.Middle},
		"small right":  {size: 100, loc: tmproofs.Right},
		"big left":     {size: 5431, loc: tmproofs.Left},
		"big middle":   {size: 5431, loc: tmproofs.Middle},
		"big right":    {size: 5431, loc: tmproofs.Right},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tree, allkeys, err := helpers.BuildTree(tc.size)
			if err != nil {
				t.Fatalf("Creating tree: %+v", err)
			}
			key := helpers.GetNonKey(allkeys, tc.loc)

			proof, err := CreateNonMembershipProof(tree, key)
			if err != nil {
				t.Fatalf("Creating Proof: %+v", err)
			}

			root := tree.WorkingHash()
			valid := ics23.VerifyNonMembership(IavlSpec, root, proof, key)
			if !valid {
				t.Fatalf("Non Membership Proof Invalid")
			}
		})
	}
}
