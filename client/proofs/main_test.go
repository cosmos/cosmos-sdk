package proofs_test

import (
	"os"
	"testing"

	"github.com/tendermint/abci/example/dummy"
	cmn "github.com/tendermint/tmlibs/common"

	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/rpc/client"
	rpctest "github.com/tendermint/tendermint/rpc/test"
)

var node *nm.Node

func getLocalClient() client.Local {
	return client.NewLocal(node)
}

func TestMain(m *testing.M) {
	// start a tendermint node (and merkleeyes) in the background to test against
	app := dummy.NewDummyApplication()
	node = rpctest.StartTendermint(app)
	code := m.Run()

	// and shut down proper at the end
	node.Stop()
	node.Wait()
	os.Exit(code)
}

func MakeTxKV() ([]byte, []byte, []byte) {
	k := cmn.RandStr(8)
	v := cmn.RandStr(8)
	tx := k + "=" + v
	return []byte(k), []byte(v), []byte(tx)
}
