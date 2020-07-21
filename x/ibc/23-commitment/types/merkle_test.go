package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
)

func (suite *MerkleTestSuite) TestVerifyMembership() {
	suite.iavlStore.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := suite.store.Commit()

	res := suite.store.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("/%s/key", suite.storeKey.Name()), // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	require.NotNil(suite.T(), res.Proof)

	proof := types.MerkleProof{
		Proof: res.Proof,
	}
	suite.Require().NoError(proof.ValidateBasic())
	suite.Require().Error(types.MerkleProof{}.ValidateBasic())

	cases := []struct {
		name       string
		root       []byte
		pathArr    []string
		value      []byte
		malleate   func()
		shouldPass bool
	}{
		{"valid proof", cid.Hash, []string{suite.storeKey.Name(), "MYKEY"}, []byte("MYVALUE"), func() {}, true},            // valid proof
		{"wrong value", cid.Hash, []string{suite.storeKey.Name(), "MYKEY"}, []byte("WRONGVALUE"), func() {}, false},        // invalid proof with wrong value
		{"nil value", cid.Hash, []string{suite.storeKey.Name(), "MYKEY"}, []byte(nil), func() {}, false},                   // invalid proof with nil value
		{"wrong key", cid.Hash, []string{suite.storeKey.Name(), "NOTMYKEY"}, []byte("MYVALUE"), func() {}, false},          // invalid proof with wrong key
		{"wrong path 1", cid.Hash, []string{suite.storeKey.Name(), "MYKEY", "MYKEY"}, []byte("MYVALUE"), func() {}, false}, // invalid proof with wrong path
		{"wrong path 2", cid.Hash, []string{suite.storeKey.Name()}, []byte("MYVALUE"), func() {}, false},                   // invalid proof with wrong path
		{"wrong path 3", cid.Hash, []string{"MYKEY"}, []byte("MYVALUE"), func() {}, false},                                 // invalid proof with wrong path
		{"wrong storekey", cid.Hash, []string{"otherStoreKey", "MYKEY"}, []byte("MYVALUE"), func() {}, false},              // invalid proof with wrong store prefix
		{"wrong root", []byte("WRONGROOT"), []string{suite.storeKey.Name(), "MYKEY"}, []byte("MYVALUE"), func() {}, false}, // invalid proof with wrong root
		{"nil root", []byte(nil), []string{suite.storeKey.Name(), "MYKEY"}, []byte("MYVALUE"), func() {}, false},           // invalid proof with nil root
		{"proof is wrong length", cid.Hash, []string{suite.storeKey.Name(), "MYKEY"}, []byte("MYVALUE"), func() {
			proof = types.MerkleProof{
				Proof: &merkle.Proof{
					Ops: res.Proof.Ops[1:],
				},
			}
		}, false}, // invalid proof with wrong length

	}

	for i, tc := range cases {
		tc := tc
		suite.Run(tc.name, func() {
			tc.malleate()

			root := types.NewMerkleRoot(tc.root)
			path := types.NewMerklePath(tc.pathArr)

			err := proof.VerifyMembership(types.GetSDKSpecs(), &root, path, tc.value)

			if tc.shouldPass {
				// nolint: scopelint
				suite.Require().NoError(err, "test case %d should have passed", i)
			} else {
				// nolint: scopelint
				suite.Require().Error(err, "test case %d should have failed", i)
			}
		})
	}

}

func (suite *MerkleTestSuite) TestVerifyNonMembership() {
	suite.iavlStore.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := suite.store.Commit()

	// Get Proof
	res := suite.store.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("/%s/key", suite.storeKey.Name()), // required path to get key/value+proof
		Data:  []byte("MYABSENTKEY"),
		Prove: true,
	})
	require.NotNil(suite.T(), res.Proof)

	proof := types.MerkleProof{
		Proof: res.Proof,
	}
	suite.Require().NoError(proof.ValidateBasic())

	cases := []struct {
		name       string
		root       []byte
		pathArr    []string
		malleate   func()
		shouldPass bool
	}{
		{"valid proof", cid.Hash, []string{suite.storeKey.Name(), "MYABSENTKEY"}, func() {}, true},            // valid proof
		{"wrong key", cid.Hash, []string{suite.storeKey.Name(), "MYKEY"}, func() {}, false},                   // invalid proof with existent key
		{"wrong path 1", cid.Hash, []string{suite.storeKey.Name(), "MYKEY", "MYABSENTKEY"}, func() {}, false}, // invalid proof with wrong path
		{"wrong path 2", cid.Hash, []string{suite.storeKey.Name(), "MYABSENTKEY", "MYKEY"}, func() {}, false}, // invalid proof with wrong path
		{"wrong path 3", cid.Hash, []string{suite.storeKey.Name()}, func() {}, false},                         // invalid proof with wrong path
		{"wrong path 4", cid.Hash, []string{"MYABSENTKEY"}, func() {}, false},                                 // invalid proof with wrong path
		{"wrong storeKey", cid.Hash, []string{"otherStoreKey", "MYABSENTKEY"}, func() {}, false},              // invalid proof with wrong store prefix
		{"wrong root", []byte("WRONGROOT"), []string{suite.storeKey.Name(), "MYABSENTKEY"}, func() {}, false}, // invalid proof with wrong root
		{"nil root", []byte(nil), []string{suite.storeKey.Name(), "MYABSENTKEY"}, func() {}, false},           // invalid proof with nil root
		{"proof is wrong length", cid.Hash, []string{suite.storeKey.Name(), "MYKEY"}, func() {
			proof = types.MerkleProof{
				Proof: &merkle.Proof{
					Ops: res.Proof.Ops[1:],
				},
			}
		}, false}, // invalid proof with wrong length

	}

	for i, tc := range cases {
		tc := tc

		suite.Run(tc.name, func() {
			tc.malleate()

			root := types.NewMerkleRoot(tc.root)
			path := types.NewMerklePath(tc.pathArr)

			err := proof.VerifyNonMembership(types.GetSDKSpecs(), &root, path)

			if tc.shouldPass {
				// nolint: scopelint
				suite.Require().NoError(err, "test case %d should have passed", i)
			} else {
				// nolint: scopelint
				suite.Require().Error(err, "test case %d should have failed", i)
			}
		})
	}

}

func TestApplyPrefix(t *testing.T) {
	prefix := types.NewMerklePrefix([]byte("storePrefixKey"))

	pathStr := "pathone/pathtwo/paththree/key"

	prefixedPath, err := types.ApplyPrefix(prefix, pathStr)
	require.Nil(t, err, "valid prefix returns error")

	require.Equal(t, "/storePrefixKey/"+pathStr, prefixedPath.Pretty(), "Prefixed path incorrect")
	require.Equal(t, "/storePrefixKey/pathone%2Fpathtwo%2Fpaththree%2Fkey", prefixedPath.String(), "Prefixed scaped path incorrect")

	// invalid prefix contains non-alphanumeric character
	invalidPathStr := "invalid-path/doesitfail?/hopefully"
	invalidPath, err := types.ApplyPrefix(prefix, invalidPathStr)
	require.NotNil(t, err, "invalid prefix does not returns error")
	require.Equal(t, types.MerklePath{}, invalidPath, "invalid prefix returns valid Path on ApplyPrefix")
}
