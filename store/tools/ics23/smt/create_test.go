package smtproofs

import (
	"crypto/sha256"
	"testing"

	ics23 "github.com/confio/ics23/go"

	tmproofs "github.com/cosmos/cosmos-sdk/store/internal/proofs"
	"github.com/cosmos/cosmos-sdk/store/tools/ics23/smt/helpers"
)

var numKeys = 50
var cases = map[string]struct {
	size int
	loc  tmproofs.Where
}{
	"tiny left":    {size: 10, loc: tmproofs.Left},
	"tiny middle":  {size: 10, loc: tmproofs.Middle},
	"tiny right":   {size: 10, loc: tmproofs.Right},
	"small left":   {size: 100, loc: tmproofs.Left},
	"small middle": {size: 100, loc: tmproofs.Middle},
	"small right":  {size: 100, loc: tmproofs.Right},
	"big left":     {size: 5431, loc: tmproofs.Left},
	"big middle":   {size: 5431, loc: tmproofs.Middle},
	"big right":    {size: 5431, loc: tmproofs.Right},
}

func TestCreateMembership(t *testing.T) {
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tree, preim, err := helpers.BuildTree(tc.size)
			if err != nil {
				t.Fatalf("Creating tree: %+v", err)
			}
			for i := 0; i < numKeys; i++ {
				key := preim.GetKey(tc.loc)
				val, err := tree.Get(key)
				if err != nil {
					t.Fatalf("Getting key: %+v", err)
				}
				proof, err := CreateMembershipProof(tree, key)
				if err != nil {
					t.Fatalf("Creating proof: %+v", err)
				}

				root := tree.Root()
				path := sha256.Sum256(key)
				valid := ics23.VerifyMembership(ics23.SmtSpec, root, proof, path[:], val)
				if !valid {
					t.Fatalf("Membership proof invalid")
				}
			}
		})
	}
}

func TestCreateNonMembership(t *testing.T) {
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tree, preim, err := helpers.BuildTree(tc.size)
			if err != nil {
				t.Fatalf("Creating tree: %+v", err)
			}

			for i := 0; i < numKeys; i++ {
				key := preim.GetNonKey(tc.loc)
				proof, err := CreateNonMembershipProof(tree, key, preim)
				if err != nil {
					t.Fatalf("Creating proof: %+v", err)
				}

				root := tree.Root()
				path := sha256.Sum256(key)
				valid := ics23.VerifyNonMembership(ics23.SmtSpec, root, proof, path[:])
				if !valid {
					t.Fatalf("Non-membership proof invalid")
				}
			}
		})
	}
}
