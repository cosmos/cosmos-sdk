package main

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/urfave/cli"

	"github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
	client "github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func cmdAccount(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("account command requires an argument ([address])")
	}
	addrHex := c.Args()[0]

	// convert destination address to bytes
	addr, err := hex.DecodeString(addrHex)
	if err != nil {
		return errors.New("Account address is invalid hex: " + err.Error())
	}

	acc, err := getAcc(c, addr)
	if err != nil {
		return err
	}
	fmt.Println(string(wire.JSONBytes(acc)))
	return nil
}

// fetch the account by querying the app
func getAcc(c *cli.Context, address []byte) (*types.Account, error) {
	tmAddr := c.String("tendermint")
	clientURI := client.NewClientURI(tmAddr)
	tmResult := new(ctypes.TMResult)

	params := map[string]interface{}{
		"path":  "/key",
		"data":  append([]byte("base/a/"), address...),
		"prove": false,
	}
	_, err := clientURI.Call("abci_query", params, tmResult)
	if err != nil {
		return nil, errors.New(cmn.Fmt("Error calling /abci_query: %v", err))
	}
	res := (*tmResult).(*ctypes.ResultABCIQuery)
	if !res.Response.Code.IsOK() {
		return nil, errors.New(cmn.Fmt("Query got non-zero exit code: %v. %s", res.Response.Code, res.Response.Log))
	}
	accountBytes := res.Response.Value

	if len(accountBytes) == 0 {
		return nil, errors.New(cmn.Fmt("Account bytes are empty from query for address %X", address))
	}
	var acc *types.Account
	err = wire.ReadBinaryBytes(accountBytes, &acc)
	if err != nil {
		return nil, errors.New(cmn.Fmt("Error reading account %X error: %v",
			accountBytes, err.Error()))
	}

	return acc, nil
}
