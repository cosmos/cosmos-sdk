package tmproofs

import (
	"testing"

	"github.com/confio/ics23-tendermint/helpers"
	ics23 "github.com/confio/ics23/go"
)

func TestCreateMembership(t *testing.T) {
	cases := map[string]struct {
		size int
		loc  helpers.Where
	}{
		"small left":   {size: 100, loc: helpers.Left},
		"small middle": {size: 100, loc: helpers.Middle},
		"small right":  {size: 100, loc: helpers.Right},
		"big left":     {size: 5431, loc: helpers.Left},
		"big middle":   {size: 5431, loc: helpers.Middle},
		"big right":    {size: 5431, loc: helpers.Right},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			data := helpers.BuildMap(tc.size)
			allkeys := helpers.SortedKeys(data)
			key := helpers.GetKey(allkeys, tc.loc)
			val := data[key]
			proof, err := CreateMembershipProof(data, []byte(key))
			if err != nil {
				t.Fatalf("Creating Proof: %+v", err)
			}
			if proof.GetExist() == nil {
				t.Fatal("Unexpected proof format")
			}

			root := helpers.CalcRoot(data)
			err = proof.GetExist().Verify(TendermintSpec, root, []byte(key), val)
			if err != nil {
				t.Fatalf("Verifying Proof: %+v", err)
			}

			valid := ics23.VerifyMembership(TendermintSpec, root, proof, []byte(key), val)
			if !valid {
				t.Fatalf("Membership Proof Invalid")
			}
		})
	}
}

func TestCreateNonMembership(t *testing.T) {
	cases := map[string]struct {
		size int
		loc  helpers.Where
	}{
		"small left":   {size: 100, loc: helpers.Left},
		"small middle": {size: 100, loc: helpers.Middle},
		"small right":  {size: 100, loc: helpers.Right},
		"big left":     {size: 5431, loc: helpers.Left},
		"big middle":   {size: 5431, loc: helpers.Middle},
		"big right":    {size: 5431, loc: helpers.Right},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			data := helpers.BuildMap(tc.size)
			allkeys := helpers.SortedKeys(data)
			key := helpers.GetNonKey(allkeys, tc.loc)

			proof, err := CreateNonMembershipProof(data, []byte(key))
			if err != nil {
				t.Fatalf("Creating Proof: %+v", err)
			}
			if proof.GetNonexist() == nil {
				t.Fatal("Unexpected proof format")
			}

			root := helpers.CalcRoot(data)
			err = proof.GetNonexist().Verify(TendermintSpec, root, []byte(key))
			if err != nil {
				t.Fatalf("Verifying Proof: %+v", err)
			}

			valid := ics23.VerifyNonMembership(TendermintSpec, root, proof, []byte(key))
			if !valid {
				t.Fatalf("Non Membership Proof Invalid")
			}
		})
	}
}
