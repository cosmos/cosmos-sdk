package tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	amino "github.com/tendermint/go-amino"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpcclient "github.com/tendermint/tendermint/rpc/lib/client"
)

// TODO: these functions just print to Stdout.
// consider using the logger.

// Uses localhost
func WaitForHeight(height int64, port string) {
	for {
		var resultBlock ctypes.ResultBlock

		url := fmt.Sprintf("http://localhost:%v%v", port, "/blocks/latest")
		res, err := http.Get(url)
		if err != nil {
			panic(err)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		res.Body.Close()

		err = cdc.UnmarshalJSON([]byte(body), &resultBlock)
		if err != nil {
			fmt.Println("RES", res)
			fmt.Println("BODY", string(body))
			panic(err)
		}

		if resultBlock.Block.Height >= height {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// wait for 2 blocks.
// uses localhost
func WaitForStart(port string) {
	waitHeight := int64(2)
	for {
		time.Sleep(time.Second)

		url := fmt.Sprintf("http://localhost:%v%v", port, "/blocks/latest")
		res, err := http.Get(url)
		if err != nil {
			panic(err)
		}

		// waiting for server to start ...
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			continue
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		res.Body.Close()

		resultBlock := new(ctypes.ResultBlock)
		err = cdc.UnmarshalJSON([]byte(body), &resultBlock)
		if err != nil {
			fmt.Println("RES", res)
			fmt.Println("BODY", string(body))
			panic(err)
		}

		if resultBlock.Block.Height >= waitHeight {
			return
		}
	}
}

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
