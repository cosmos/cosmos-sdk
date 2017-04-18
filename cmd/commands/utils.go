package commands

import (
	"encoding/hex"
	"errors"
	"os"

	"github.com/urfave/cli"

	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/go-common"
	client "github.com/tendermint/go-rpc/client"
	wire "github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func BasecoinRoot(rootDir string) string {
	if rootDir == "" {
		rootDir = os.Getenv("BCHOME")
	}
	if rootDir == "" {
		rootDir = os.Getenv("HOME") + "/.basecoin"
	}
	return rootDir
}

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

func StripHex(s string) string {
	if isHex(s) {
		return s[2:]
	}
	return s
}

func Query(tmAddr string, key []byte) (*abci.ResponseQuery, error) {
	uriClient := client.NewURIClient(tmAddr)
	tmResult := new(ctypes.TMResult)

	params := map[string]interface{}{
		"path":  "/key",
		"data":  key,
		"prove": true,
	}
	_, err := uriClient.Call("abci_query", params, tmResult)
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

	key := state.AccountKey(address)
	response, err := Query(tmAddr, key)
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

func getHeaderAndCommit(c *cli.Context, height int) (*tmtypes.Header, *tmtypes.Commit, error) {
	tmResult := new(ctypes.TMResult)
	tmAddr := c.String("node")
	uriClient := client.NewURIClient(tmAddr)

	method := "commit"
	_, err := uriClient.Call(method, map[string]interface{}{"height": height}, tmResult)
	if err != nil {
		return nil, nil, errors.New(cmn.Fmt("Error on %s: %v", method, err))
	}
	resCommit := (*tmResult).(*ctypes.ResultCommit)
	header := resCommit.Header
	commit := resCommit.Commit

	return header, commit, nil
}
