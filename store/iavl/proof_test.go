package iavl

import (
	"testing"

	ics23 "github.com/confio/ics23/go"
	"github.com/stretchr/testify/require"
)

func TestConvertExistence(t *testing.T) {
	proof, err := GenerateResult(200, Middle)
	require.NoError(t, err)

	converted, err := convertExistenceProof(proof.Proof, proof.Key, proof.Value)
	require.NoError(t, err)

	calc, err := converted.Calculate()
	require.NoError(t, err)

	require.Equal(t, []byte(calc), proof.RootHash, "Calculated: %X\nExpected:   %X", calc, proof.RootHash)
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
			require.NoError(t, err, "Creating tree: %+v", err)

			key := GetKey(allkeys, tc.loc)
			_, val := tree.Get(key)
			proof, err := CreateMembershipProof(tree, key)
			require.NoError(t, err, "Creating Proof: %+v", err)

			root := tree.Hash()
			valid := ics23.VerifyMembership(ics23.IavlSpec, root, proof, key, val)
			if !valid {
				require.NoError(t, err, "Membership Proof Invalid")
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
			require.NoError(t, err, "Creating tree: %+v", err)

			key := GetNonKey(allkeys, tc.loc)

			proof, err := CreateNonMembershipProof(tree, key)
			require.NoError(t, err, "Creating Proof: %+v", err)

			root := tree.Hash()
			valid := ics23.VerifyNonMembership(ics23.IavlSpec, root, proof, key)
			if !valid {
				require.NoError(t, err, "Non Membership Proof Invalid")
			}
		})
	}
}
