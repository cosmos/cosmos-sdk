package tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Wait for N tendermint blocks to pass using the Tendermint RPC
// on localhost
func WaitForNextNBlocksTM(n int64, port string) {
	// get the latest block and wait for n more
	url := fmt.Sprintf("http://localhost:%v", port)
	cl, err := rpchttp.New(url, "/websocket")
	if err != nil {
		panic(fmt.Sprintf("failed to create Tendermint HTTP client: %s", err))
	}

	var height int64

	resBlock, err := cl.Block(nil)
	if err != nil || resBlock.Block == nil {
		// wait for the first block to exist
		WaitForHeightTM(1, port)
		height = 1 + n
	} else {
		height = resBlock.Block.Height + n
	}

	waitForHeightTM(height, url)
}

// Wait for the given height from the Tendermint RPC
// on localhost
func WaitForHeightTM(height int64, port string) {
	url := fmt.Sprintf("http://localhost:%v", port)
	waitForHeightTM(height, url)
}

func waitForHeightTM(height int64, url string) {
	cl, err := rpchttp.New(url, "/websocket")
	if err != nil {
		panic(fmt.Sprintf("failed to create Tendermint HTTP client: %s", err))
	}

	for {
		// get url, try a few times
		var resBlock *ctypes.ResultBlock
		var err error
	INNER:
		for i := 0; i < 5; i++ {
			resBlock, err = cl.Block(nil)
			if err == nil {
				break INNER
			}
			time.Sleep(time.Millisecond * 200)
		}
		if err != nil {
			panic(err)
		}

		if resBlock.Block != nil && resBlock.Block.Height >= height {
			return
		}

		time.Sleep(time.Millisecond * 100)
	}
}

// wait for tendermint to start by querying tendermint
func WaitForTMStart(port string) {
	url := fmt.Sprintf("http://localhost:%v/block", port)
	WaitForStart(url)
}

// WaitForStart waits for the node to start by pinging the url
// every 100ms for 10s until it returns 200. If it takes longer than 5s,
// it panics.
func WaitForStart(url string) {
	var err error

	// ping the status endpoint a few times a second
	// for a few seconds until we get a good response.
	// otherwise something probably went wrong
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)

		var res *http.Response
		res, err = http.Get(url) // nolint:gosec
		if err != nil || res == nil {
			continue
		}
		//		body, _ := ioutil.ReadAll(res.Body)
		//		fmt.Println("BODY", string(body))
		err = res.Body.Close()
		if err != nil {
			panic(err)
		}

		if res.StatusCode == http.StatusOK {
			// good!
			return
		}
	}
	// still haven't started up?! panic!
	panic(err)
}

// NewTestCaseDir creates a new temporary directory for a test case.
// Returns the directory path and a cleanup function.
// nolint: errcheck
func NewTestCaseDir(t testing.TB) (string, func()) {
	dir, err := ioutil.TempDir("", t.Name()+"_")
	require.NoError(t, err)
	return dir, func() { os.RemoveAll(dir) }
}

var cdc = codec.New()

func init() {
	ctypes.RegisterAmino(cdc.Amino)
}

//DONTCOVER
