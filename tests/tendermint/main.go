package main

import (
	"encoding/hex"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/tendermint/blackstar/tests"
	"github.com/tendermint/blackstar/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-rpc/types"
	"github.com/tendermint/go-wire"
	_ "github.com/tendermint/tendermint/rpc/core/types" // Register RPCResponse > Result types
)

func main() {
	ws := rpcclient.NewWSClient("ws://127.0.0.1:46657/websocket")
	// ws := rpcclient.NewWSClient("ws://104.236.69.128:46657/websocket")
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
			fmt.Println("res:", Blue(string(res)))
		}
	}()

	// Get the root account
	root := tests.PrivAccountFromSecret("root")
	sequence := uint(0)
	// Make a bunch of PrivAccounts
	privAccounts := tests.RandAccounts(1000, 1000000, 0)

	// Send coins to each account
	for i := 0; i < len(privAccounts); i++ {
		privAccount := privAccounts[i]
		tx := types.Tx{
			Inputs: []types.Input{
				types.Input{
					PubKey:   root.PubKey,
					Amount:   1000002,
					Sequence: sequence,
				},
			},
			Outputs: []types.Output{
				types.Output{
					PubKey: privAccount.PubKey,
					Amount: 1000000,
				},
			},
		}
		sequence += 1

		// Write request
		txBytes := wire.BinaryBytes(tx)
		fmt.Println("tx:", hex.EncodeToString(txBytes))
		request := rpctypes.NewRPCRequest("fakeid", "broadcast_tx", Arr(txBytes))
		reqBytes := wire.JSONBytes(request)
		fmt.Println("req:", hex.EncodeToString(reqBytes))
		//fmt.Print(".")
		err := ws.WriteMessage(websocket.TextMessage, reqBytes)
		if err != nil {
			Exit("writing websocket request: " + err.Error())
		}
	}

	/*
		// Make a bunch of requests
		for i := 0; ; i++ {
			binary.BigEndian.PutUint64(buf, uint64(i))
			//txBytes := hex.EncodeToString(buf[:n])
			request := rpctypes.NewRPCRequest("fakeid", "broadcast_tx", Arr(buf[:8]))
			reqBytes := wire.JSONBytes(request)
			//fmt.Println("!!", string(reqBytes))
			fmt.Print(".")
			err := ws.WriteMessage(websocket.TextMessage, reqBytes)
			if err != nil {
				Exit(err.Error())
			}
			if i%1000 == 0 {
				fmt.Println(i)
			}
			time.Sleep(time.Microsecond * 1000)
		}
	*/

	ws.Stop()
}
