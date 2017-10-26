package proofs_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ctest "github.com/tendermint/tmlibs/test"

	"github.com/tendermint/tendermint/certifiers"
	certcli "github.com/tendermint/tendermint/certifiers/client"
	"github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client/proofs"
)

func getCurrentCheck(t *testing.T, cl client.Client) certifiers.Commit {
	stat, err := cl.Status()
	require.Nil(t, err, "%+v", err)
	return getCheckForHeight(t, cl, stat.LatestBlockHeight)
}

func getCheckForHeight(t *testing.T, cl client.Client, h int) certifiers.Commit {
	client.WaitForHeight(cl, h, nil)
	commit, err := cl.Commit(&h)
	require.Nil(t, err, "%+v", err)
	return certcli.CommitFromResult(commit)
}

func TestAppProofs(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cl := getLocalClient()
	prover := proofs.NewAppProver(cl)
	time.Sleep(200 * time.Millisecond)

	precheck := getCurrentCheck(t, cl)

	// great, let's store some data here, and make more checks....
	k, v, tx := MakeTxKV()
	br, err := cl.BroadcastTxCommit(tx)
	require.Nil(err, "%+v", err)
	require.EqualValues(0, br.CheckTx.Code)
	require.EqualValues(0, br.DeliverTx.Code)
	h := br.Height + 1

	// unfortunately we cannot tell the server to give us any height
	// other than the most recent, so 0 is the only choice :(
	pr, err := prover.Get(k, uint64(h))
	require.Nil(err, "%+v", err)
	check := getCheckForHeight(t, cl, h)

	// matches and validates with post-tx header
	err = pr.Validate(check)
	assert.Nil(err, "%+v", err)

	// doesn't matches with pre-tx header
	err = pr.Validate(precheck)
	assert.NotNil(err)

	// make sure it has the values we want
	apk, ok := pr.(proofs.AppProof)
	if assert.True(ok) {
		assert.EqualValues(k, apk.Key)
		assert.EqualValues(v, apk.Value)
	}

	// make sure we read/write properly, and any changes to the serialized
	// object are invalid proof (2000 random attempts)

	// TODO: iavl panics, fix this
	// testSerialization(t, prover, pr, check, 2000)
}

// testSerialization makes sure the proof will un/marshal properly
// and validate with the checkpoint.  It also does lots of modifications
// to the binary data and makes sure no mods validates properly
func testSerialization(t *testing.T, prover proofs.Prover, pr proofs.Proof,
	check certifiers.Commit, mods int) {

	require := require.New(t)

	// first, make sure that we can serialize and deserialize
	err := pr.Validate(check)
	require.Nil(err, "%+v", err)

	// store the data
	data, err := pr.Marshal()
	require.Nil(err, "%+v", err)

	// recover the data and make sure it still checks out
	npr, err := prover.Unmarshal(data)
	require.Nil(err, "%+v", err)
	err = npr.Validate(check)
	require.Nil(err, "%#v\n%+v", npr, err)

	// now let's go mod...
	for i := 0; i < mods; i++ {
		bdata := ctest.MutateByteSlice(data)
		bpr, err := prover.Unmarshal(bdata)
		if err == nil {
			assert.NotNil(t, bpr.Validate(check))
		}
	}
}

// // validate all tx in the block
// block, err := cl.Block(check.Height())
// require.Nil(err, "%+v", err)
// err = check.CheckTxs(block.Block.Data.Txs)
// assert.Nil(err, "%+v", err)

// oh, i would like the know the hieght of the broadcast_commit.....
// so i could verify that tx :(
