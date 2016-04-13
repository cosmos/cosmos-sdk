// +build scripts

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-rpc/types"
	"github.com/tendermint/go-wire"
	_ "github.com/tendermint/tendermint/rpc/core/types" // Register RPCResponse > Result types
)

func main() {
	ws := rpcclient.NewWSClient(os.Args[1]+":46657", "/websocket")

	_, err := ws.Start()
	if err != nil {
		Exit(err.Error())
	}

	// Read a bunch of responses
	go func() {
		for {
			res, ok := <-ws.ResultsCh
			if !ok {
				break
			}
			//fmt.Println(counter, "res:", Blue(string(res)))
			var result []interface{}
			err := json.Unmarshal([]byte(string(res)), &result)
			if err != nil {
				Exit("Error unmarshalling block: " + err.Error())
			}
			height := result[1].(map[string]interface{})["block"].(map[string]interface{})["header"].(map[string]interface{})["height"]
			txs := result[1].(map[string]interface{})["block"].(map[string]interface{})["data"].(map[string]interface{})["txs"]
			if len(txs.([]interface{})) > 0 {
				fmt.Println(">>", height, txs)
			}
		}
	}()

	for i := 0; i < 100000; i++ {
		request := rpctypes.NewRPCRequest("fakeid", "block", Arr(i))
		reqBytes := wire.JSONBytes(request)
		err = ws.WriteMessage(websocket.TextMessage, reqBytes)
		if err != nil {
			Exit("writing websocket request: " + err.Error())
		}
	}

	time.Sleep(time.Second * 1000)

	ws.Stop()
}
