package main

import (
	"encoding/hex"
	"errors"

	"github.com/urfave/cli"

	"github.com/tendermint/basecoin/types"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/go-common"
	client "github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Returns true for non-empty hex-string prefixed with "0x"
func isHex(s string) bool {
	if len(s) > 2 && s[:2] == "0x" {
		_, err := hex.DecodeString(s[2:])
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func stripHex(s string) string {
	if isHex(s) {
		return s[2:]
	}
	return s
}

func query(tmAddr string, key []byte) (*abci.ResponseQuery, error) {
	clientURI := client.NewClientURI(tmAddr)
	tmResult := new(ctypes.TMResult)

	params := map[string]interface{}{
		"path":  "/key",
		"data":  key,
		"prove": true,
	}
	_, err := clientURI.Call("abci_query", params, tmResult)
	if err != nil {
		return nil, errors.New(cmn.Fmt("Error calling /abci_query: %v", err))
	}
	res := (*tmResult).(*ctypes.ResultABCIQuery)
	if !res.Response.Code.IsOK() {
		return nil, errors.New(cmn.Fmt("Query got non-zero exit code: %v. %s", res.Response.Code, res.Response.Log))
	}
	return &res.Response, nil
}

// fetch the account by querying the app
func getAcc(tmAddr string, address []byte) (*types.Account, error) {

	key := append([]byte("base/a/"), address...)
	response, err := query(tmAddr, key)
	if err != nil {
		return nil, err
	}

	accountBytes := response.Value

	if len(accountBytes) == 0 {
		return nil, errors.New(cmn.Fmt("Account bytes are empty for address: %X ", address))
	}

	var acc *types.Account
	err = wire.ReadBinaryBytes(accountBytes, &acc)
	if err != nil {
		return nil, errors.New(cmn.Fmt("Error reading account %X error: %v",
			accountBytes, err.Error()))
	}

	return acc, nil
}

func getBlock(c *cli.Context, height int) (*tmtypes.Block, error) {
	tmResult := new(ctypes.TMResult)
	tmAddr := c.String("node")
	clientURI := client.NewClientURI(tmAddr)

	_, err := clientURI.Call("block", map[string]interface{}{"height": height}, tmResult)
	if err != nil {
		return nil, errors.New(cmn.Fmt("Error on broadcast tx: %v", err))
	}
	res := (*tmResult).(*ctypes.ResultBlock)
	return res.Block, nil
}
