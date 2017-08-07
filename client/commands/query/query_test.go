package query

import (
	"os"
	"testing"
	"time"

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
	time.Sleep(200 * time.Millisecond)

	k := []byte("my-key")
	v := []byte("my-value")

	tx := eyes.SetTx{Key: k, Value: v}.Wrap()
	btx := wire.BinaryBytes(tx)
	br, err := cl.BroadcastTxCommit(btx)
	require.Nil(err, "%+v", err)
	require.EqualValues(0, br.CheckTx.Code, "%#v", br.CheckTx)
	require.EqualValues(0, br.DeliverTx.Code)

	// this sets up our trust on the node based on some past point.
	// maybe this can be cleaned up and made easy to reuse
	source := certclient.New(cl)
	trusted := certifiers.NewMemStoreProvider()
	// let's start with some trust before the query...
	seed, err := source.GetByHeight(br.Height - 2)
	require.Nil(err, "%+v", err)
	cert := certifiers.NewInquiring("my-chain", seed, trusted, source)

	// Test existing key.

	bs, _, proof, err := CustomGetWithProof(k, cl, cert)
	require.Nil(err, "%+v", err)
	require.NotNil(proof)

	var data eyes.Data
	err = wire.ReadBinaryBytes(bs, &data)
	require.Nil(err, "%+v", err)
	assert.EqualValues(v, data.Value)

	// Test non-existing key.

	// TODO: This currently fails.
	missing := []byte("my-missing-key")
	bs, _, proof, err = CustomGetWithProof(missing, cl, cert)
	require.Nil(err, "%+v", err)
	require.Nil(bs)
	require.NotNil(proof)
}
