package tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	amino "github.com/tendermint/go-amino"
	tmclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpcclient "github.com/tendermint/tendermint/rpc/lib/client"
)

// Wait for the next tendermint block from the Tendermint RPC
// on localhost
func WaitForNextHeightTM(port string) {
	url := fmt.Sprintf("http://localhost:%v", port)
	cl := tmclient.NewHTTP(url, "/websocket")
	resBlock, err := cl.Block(nil)
	if err != nil {
		panic(err)
	}
	waitForHeightTM(resBlock.Block.Height+1, url)
}

// Wait for the given height from the Tendermint RPC
// on localhost
func WaitForHeightTM(height int64, port string) {
	url := fmt.Sprintf("http://localhost:%v", port)
	waitForHeightTM(height, url)
}

func waitForHeightTM(height int64, url string) {
	cl := tmclient.NewHTTP(url, "/websocket")
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

		if resBlock.Block != nil &&
			resBlock.Block.Height >= height {
			fmt.Println("HEIGHT", resBlock.Block.Height)
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// Wait for height from the LCD API on localhost
func WaitForHeight(height int64, port string) {
	url := fmt.Sprintf("http://localhost:%v/blocks/latest", port)
	waitForHeight(height, url)
}

func waitForHeight(height int64, url string) {
	for {
		// get url, try a few times
		var res *http.Response
		var err error
		for i := 0; i < 5; i++ {
			res, err = http.Get(url)
			if err == nil {
				break
			}
			time.Sleep(time.Millisecond * 200)
		}
		if err != nil {
			panic(err)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		res.Body.Close()

		var resultBlock ctypes.ResultBlock
		err = cdc.UnmarshalJSON([]byte(body), &resultBlock)
		if err != nil {
			fmt.Println("RES", res)
			fmt.Println("BODY", string(body))
			panic(err)
		}

		if resultBlock.Block != nil &&
			resultBlock.Block.Height >= height {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// wait for tendermint to start
func WaitForStart(port string) {
	var err error
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)

		url := fmt.Sprintf("http://localhost:%v/blocks/latest", port)

		// get url, try a few times
		var res *http.Response
		res, err = http.Get(url)
		if err == nil || res == nil {
			continue
		}

		// waiting for server to start ...
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			return
		}
	}
	if err != nil {
		panic(err)
	}
}

// TODO: these functions just print to Stdout.
// consider using the logger.

// Wait for the RPC server to respond to /status
func WaitForRPC(laddr string) {
	fmt.Println("LADDR", laddr)
	client := rpcclient.NewJSONRPCClient(laddr)
	ctypes.RegisterAmino(client.Codec())
	result := new(ctypes.ResultStatus)
	for {
		_, err := client.Call("status", map[string]interface{}{}, result)
		if err == nil {
			return
		}
		fmt.Printf("Waiting for RPC server to start on %s:%v\n", laddr, err)
		time.Sleep(time.Millisecond)
	}
}

var cdc = amino.NewCodec()

func init() {
	ctypes.RegisterAmino(cdc)
}
