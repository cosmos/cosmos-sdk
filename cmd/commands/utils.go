package commands

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/go-common"
	client "github.com/tendermint/go-rpc/client"
	wire "github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

//This variable can be overwritten by plugin applications
// if they require a different working directory
var DefaultHome = ".basecoin"

func BasecoinRoot(rootDir string) string {
	if rootDir == "" {
		rootDir = os.Getenv("BCHOME")
	}
	if rootDir == "" {
		rootDir = path.Join(os.Getenv("HOME"), DefaultHome)
	}
	return rootDir
}

//Add debugging flag and execute the root command
func ExecuteWithDebug(RootCmd *cobra.Command) {

	var debug bool
	RootCmd.SilenceUsage = true
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enables stack trace error messages")

	//note that Execute() prints the error if encountered, so no need to reprint the error,
	//  only if we want the full stack trace
	if err := RootCmd.Execute(); err != nil && debug {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}
}

//Quickly registering flags can be quickly achieved through using the utility functions
//RegisterFlags, and RegisterPersistentFlags. Ex:
//	flags := []Flag2Register{
//		{&myStringFlag, "mystringflag", "foobar", "description of what this flag does"},
//		{&myBoolFlag, "myboolflag", false, "description of what this flag does"},
//		{&myInt64Flag, "myintflag", 333, "description of what this flag does"},
//	}
//	RegisterFlags(MyCobraCmd, flags)
type Flag2Register struct {
	Pointer interface{}
	Use     string
	Value   interface{}
	Desc    string
}

//register flag utils
func RegisterFlags(c *cobra.Command, flags []Flag2Register) {
	registerFlags(c, flags, false)
}

func RegisterPersistentFlags(c *cobra.Command, flags []Flag2Register) {
	registerFlags(c, flags, true)
}

func registerFlags(c *cobra.Command, flags []Flag2Register, persistent bool) {

	var flagset *pflag.FlagSet
	if persistent {
		flagset = c.PersistentFlags()
	} else {
		flagset = c.Flags()
	}

	for _, f := range flags {

		ok := false

		switch f.Value.(type) {
		case string:
			if _, ok = f.Pointer.(*string); ok {
				flagset.StringVar(f.Pointer.(*string), f.Use, f.Value.(string), f.Desc)
			}
		case int:
			if _, ok = f.Pointer.(*int); ok {
				flagset.IntVar(f.Pointer.(*int), f.Use, f.Value.(int), f.Desc)
			}
		case uint64:
			if _, ok = f.Pointer.(*uint64); ok {
				flagset.Uint64Var(f.Pointer.(*uint64), f.Use, f.Value.(uint64), f.Desc)
			}
		case bool:
			if _, ok = f.Pointer.(*bool); ok {
				flagset.BoolVar(f.Pointer.(*bool), f.Use, f.Value.(bool), f.Desc)
			}
		}

		if !ok {
			panic("improper use of RegisterFlags")
		}
	}
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
		return nil, errors.Errorf("Error calling /abci_query: %v", err)
	}
	res := (*tmResult).(*ctypes.ResultABCIQuery)
	if !res.Response.Code.IsOK() {
		return nil, errors.Errorf("Query got non-zero exit code: %v. %s", res.Response.Code, res.Response.Log)
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
		return nil, fmt.Errorf("Account bytes are empty for address: %X ", address) //never stack trace
	}

	var acc *types.Account
	err = wire.ReadBinaryBytes(accountBytes, &acc)
	if err != nil {
		return nil, errors.Errorf("Error reading account %X error: %v",
			accountBytes, err.Error())
	}

	return acc, nil
}

func getHeaderAndCommit(tmAddr string, height int) (*tmtypes.Header, *tmtypes.Commit, error) {
	tmResult := new(ctypes.TMResult)
	uriClient := client.NewURIClient(tmAddr)

	method := "commit"
	_, err := uriClient.Call(method, map[string]interface{}{"height": height}, tmResult)
	if err != nil {
		return nil, nil, errors.Errorf("Error on %s: %v", method, err)
	}
	resCommit := (*tmResult).(*ctypes.ResultCommit)
	header := resCommit.Header
	commit := resCommit.Commit

	return header, commit, nil
}
