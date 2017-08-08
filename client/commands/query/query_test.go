package query

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/go-wire"
	"github.com/tendermint/light-client/certifiers"
	certclient "github.com/tendermint/light-client/certifiers/client"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/rpc/client"
	rpctest "github.com/tendermint/tendermint/rpc/test"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/modules/eyes"
)

var node *nm.Node

func TestMain(m *testing.M) {
	logger := log.TestingLogger()
	store, err := app.NewStore("", 0, logger)
	if err != nil {
		panic(err)
	}
	app := app.NewBasecoin(eyes.NewHandler(), store, logger)
	node = rpctest.StartTendermint(app)

	code := m.Run()

	node.Stop()
	node.Wait()
	os.Exit(code)
}

func TestAppProofs(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cl := client.NewLocal(node)
	client.WaitForHeight(cl, 1, nil)

	k := []byte("my-key")
	v := []byte("my-value")

	tx := eyes.SetTx{Key: k, Value: v}.Wrap()
	btx := wire.BinaryBytes(tx)
	br, err := cl.BroadcastTxCommit(btx)
	require.Nil(err, "%+v", err)
	require.EqualValues(0, br.CheckTx.Code, "%#v", br.CheckTx)
	require.EqualValues(0, br.DeliverTx.Code)

	// This sets up our trust on the node based on some past point.
	source := certclient.New(cl)
	seed, err := source.GetByHeight(br.Height - 2)
	require.Nil(err, "%+v", err)
	cert := certifiers.NewStatic("my-chain", seed.Validators)

	// Test existing key.
	var data eyes.Data

	bs, _, proofExists, _, err := getWithProof(k, cl, cert)
	require.Nil(err, "%+v", err)
	require.NotNil(proofExists)

	err = wire.ReadBinaryBytes(bs, &data)
	require.Nil(err, "%+v", err)
	assert.EqualValues(v, data.Value)
	err = proofExists.Verify(k, bs, proofExists.RootHash)
	assert.Nil(err, "%+v", err)

	// Test non-existing key.
	missing := []byte("my-missing-key")
	bs, _, proofExists, proofNotExists, err := getWithProof(missing, cl, cert)
	require.Nil(err, "%+v", err)
	require.Nil(bs)
	require.Nil(proofExists)
	require.NotNil(proofNotExists)
	err = proofNotExists.Verify(missing, proofNotExists.RootHash)
	assert.Nil(err, "%+v", err)
	err = proofNotExists.Verify(k, proofNotExists.RootHash)
	assert.NotNil(err)
}
