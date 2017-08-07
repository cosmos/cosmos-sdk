package query

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/modules/etc"
	"github.com/tendermint/go-wire"
	lc "github.com/tendermint/light-client"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/rpc/client"
	rpctest "github.com/tendermint/tendermint/rpc/test"
	"github.com/tendermint/tmlibs/log"
	ctest "github.com/tendermint/tmlibs/test"
)

var node *nm.Node

func getLocalClient() client.Local {
	return client.NewLocal(node)
}

func TestMain(m *testing.M) {
	// start a tendermint node (and merkleeyes) in the background to test against
	logger := log.TestingLogger()
	store, err := app.NewStore("", 0, logger)
	if err != nil {
		panic(err)
	}
	app := app.NewBasecoin(etc.NewHandler(), store, logger)
	node = rpctest.StartTendermint(app)
	code := m.Run()

	// and shut down proper at the end
	node.Stop()
	node.Wait()
	os.Exit(code)
}

func getCurrentCheck(t *testing.T, cl client.Client) lc.Checkpoint {
	stat, err := cl.Status()
	require.Nil(t, err, "%+v", err)
	return getCheckForHeight(t, cl, stat.LatestBlockHeight)
}

func getCheckForHeight(t *testing.T, cl client.Client, h int) lc.Checkpoint {
	client.WaitForHeight(cl, h, nil)
	commit, err := cl.Commit(h)
	require.Nil(t, err, "%+v", err)
	return lc.CheckpointFromResult(commit)
}

func TestAppProofs(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cl := getLocalClient()
	// prover := proofs.NewAppProver(cl)
	time.Sleep(200 * time.Millisecond)

	// precheck := getCurrentCheck(t, cl)

	k := []byte("my-key")
	v := []byte("my-value")
	// var height uint64 = 123

	// great, let's store some data here, and make more checks....
	// k, v, tx := merktest.MakeTxKV()
	tx := etc.SetTx{Key: k, Value: v}.Wrap()
	btx := wire.BinaryBytes(tx)
	br, err := cl.BroadcastTxCommit(btx)
	require.Nil(err, "%+v", err)
	require.EqualValues(0, br.CheckTx.Code, "%#v", br.CheckTx)
	require.EqualValues(0, br.DeliverTx.Code)

	// unfortunately we cannot tell the server to give us any height
	// other than the most recent, so 0 is the only choice :(
	val, _, proof, err := GetWithProof(k)
	require.Nil(err, "%+v", err)
	require.NotNil(proof)
	assert.EqualValues(v, val)
	// check := getCheckForHeight(t, cl, int(h))

	// matches and validates with post-tx header
	// err = pr.Verify(check)
	// assert.Nil(err, "%+v", err)

	// doesn't matches with pre-tx header
	// err = pr.Validate(precheck)
	// assert.NotNil(err)

	// make sure we read/write properly, and any changes to the serialized
	// object are invalid proof (2000 random attempts)
	// testSerialization(t, prover, pr, check, 2000)
}

// testSerialization makes sure the proof will un/marshal properly
// and validate with the checkpoint.  It also does lots of modifications
// to the binary data and makes sure no mods validates properly
func testSerialization(t *testing.T, prover lc.Prover, pr lc.Proof,
	check lc.Checkpoint, mods int) {

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
