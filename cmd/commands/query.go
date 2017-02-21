package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/urfave/cli"

	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-merkle"
	"github.com/tendermint/go-wire"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	QueryCmd = cli.Command{
		Name:      "query",
		Usage:     "Query the merkle tree",
		ArgsUsage: "<key>",
		Action: func(c *cli.Context) error {
			return cmdQuery(c)
		},
		Flags: []cli.Flag{
			NodeFlag,
		},
	}

	AccountCmd = cli.Command{
		Name:      "account",
		Usage:     "Get details of an account",
		ArgsUsage: "<address>",
		Action: func(c *cli.Context) error {
			return cmdAccount(c)
		},
		Flags: []cli.Flag{
			NodeFlag,
		},
	}

	BlockCmd = cli.Command{
		Name:      "block",
		Usage:     "Get the header and commit of a block",
		ArgsUsage: "<height>",
		Action: func(c *cli.Context) error {
			return cmdBlock(c)
		},
		Flags: []cli.Flag{
			NodeFlag,
		},
	}

	VerifyCmd = cli.Command{
		Name:  "verify",
		Usage: "Verify the IAVL proof",
		Action: func(c *cli.Context) error {
			return cmdVerify(c)
		},
		Flags: []cli.Flag{
			ProofFlag,
			KeyFlag,
			ValueFlag,
			RootFlag,
		},
	}
)

// Register a subcommand of QueryCmd for plugin specific query functionality
func RegisterQuerySubcommand(cmd cli.Command) {
	QueryCmd.Subcommands = append(QueryCmd.Subcommands, cmd)
}

func cmdQuery(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("query command requires an argument ([key])")
	}
	keyString := c.Args()[0]
	key := []byte(keyString)
	if isHex(keyString) {
		// convert key to bytes
		var err error
		key, err = hex.DecodeString(StripHex(keyString))
		if err != nil {
			return errors.New(cmn.Fmt("Query key (%v) is invalid hex: %v", keyString, err))
		}
	}

	resp, err := Query(c.String("node"), key)
	if err != nil {
		return err
	}

	if !resp.Code.IsOK() {
		return errors.New(cmn.Fmt("Query for key (%v) returned non-zero code (%v): %v", keyString, resp.Code, resp.Log))
	}

	val := resp.Value
	proof := resp.Proof
	height := resp.Height

	fmt.Println(string(wire.JSONBytes(struct {
		Value  []byte `json:"value"`
		Proof  []byte `json:"proof"`
		Height uint64 `json:"height"`
	}{val, proof, height})))

	return nil
}

func cmdAccount(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("account command requires an argument ([address])")
	}
	addrHex := StripHex(c.Args()[0])

	// convert destination address to bytes
	addr, err := hex.DecodeString(addrHex)
	if err != nil {
		return errors.New(cmn.Fmt("Account address (%v) is invalid hex: %v", addrHex, err))
	}

	acc, err := getAcc(c.String("node"), addr)
	if err != nil {
		return err
	}
	fmt.Println(string(wire.JSONBytes(acc)))
	return nil
}

func cmdBlock(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("block command requires an argument ([height])")
	}
	heightString := c.Args()[0]
	height, err := strconv.Atoi(heightString)
	if err != nil {
		return errors.New(cmn.Fmt("Height must be an int, got %v: %v", heightString, err))
	}

	header, commit, err := getHeaderAndCommit(c, height)
	if err != nil {
		return err
	}

	fmt.Println(string(wire.JSONBytes(struct {
		Hex  BlockHex  `json:"hex"`
		JSON BlockJSON `json:"json"`
	}{
		BlockHex{
			Header: wire.BinaryBytes(header),
			Commit: wire.BinaryBytes(commit),
		},
		BlockJSON{
			Header: header,
			Commit: commit,
		},
	})))

	return nil
}

type BlockHex struct {
	Header []byte `json:"header"`
	Commit []byte `json:"commit"`
}

type BlockJSON struct {
	Header *tmtypes.Header `json:"header"`
	Commit *tmtypes.Commit `json:"commit"`
}

func cmdVerify(c *cli.Context) error {
	keyString, valueString := c.String("key"), c.String("value")

	var err error
	key := []byte(keyString)
	if isHex(keyString) {
		key, err = hex.DecodeString(StripHex(keyString))
		if err != nil {
			return errors.New(cmn.Fmt("Key (%v) is invalid hex: %v", keyString, err))
		}
	}

	value := []byte(valueString)
	if isHex(valueString) {
		value, err = hex.DecodeString(StripHex(valueString))
		if err != nil {
			return errors.New(cmn.Fmt("Value (%v) is invalid hex: %v", valueString, err))
		}
	}

	root, err := hex.DecodeString(StripHex(c.String("root")))
	if err != nil {
		return errors.New(cmn.Fmt("Root (%v) is invalid hex: %v", c.String("root"), err))
	}

	proofBytes, err := hex.DecodeString(StripHex(c.String("proof")))
	if err != nil {
		return errors.New(cmn.Fmt("Proof (%v) is invalid hex: %v", c.String("proof"), err))
	}

	proof, err := merkle.ReadProof(proofBytes)
	if err != nil {
		return errors.New(cmn.Fmt("Error unmarshalling proof: %v", err))
	}

	if proof.Verify(key, value, root) {
		fmt.Println("OK")
	} else {
		return errors.New("Proof does not verify")
	}
	return nil
}
