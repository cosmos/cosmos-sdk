package commands

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tendermint/go-merkle"
	"github.com/tendermint/go-wire"
	tmtypes "github.com/tendermint/tendermint/types"
)

//commands
var (
	QueryCmd = &cobra.Command{
		Use:   "query [key]",
		Short: "Query the merkle tree",
		RunE:  queryCmd,
	}

	AccountCmd = &cobra.Command{
		Use:   "account [address]",
		Short: "Get details of an account",
		RunE:  accountCmd,
	}

	BlockCmd = &cobra.Command{
		Use:   "block [height]",
		Short: "Get the header and commit of a block",
		RunE:  blockCmd,
	}

	VerifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "Verify the IAVL proof",
		RunE:  verifyCmd,
	}
)

//flags
var (
	nodeFlag  string
	proofFlag string
	keyFlag   string
	valueFlag string
	rootFlag  string
)

func init() {

	commonFlags := []Flag2Register{
		{&nodeFlag, "node", "tcp://localhost:46657", "Tendermint RPC address"},
	}

	verifyFlags := []Flag2Register{
		{&proofFlag, "proof", "", "hex-encoded IAVL proof"},
		{&keyFlag, "key", "", "key to the IAVL tree"},
		{&valueFlag, "value", "", "value in the IAVL tree"},
		{&rootFlag, "root", "", "root hash of the IAVL tree"},
	}

	RegisterFlags(QueryCmd, commonFlags)
	RegisterFlags(AccountCmd, commonFlags)
	RegisterFlags(BlockCmd, commonFlags)
	RegisterFlags(VerifyCmd, verifyFlags)
}

func queryCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return fmt.Errorf("query command requires an argument ([key])") //never stack trace
	}

	keyString := args[0]
	key := []byte(keyString)
	if isHex(keyString) {
		// convert key to bytes
		var err error
		key, err = hex.DecodeString(StripHex(keyString))
		if err != nil {
			return errors.Errorf("Query key (%v) is invalid hex: %v\n", keyString, err)
		}
	}

	resp, err := Query(nodeFlag, key)
	if err != nil {
		return errors.Errorf("Query returns error: %v\n", err)
	}

	if !resp.Code.IsOK() {
		return errors.Errorf("Query for key (%v) returned non-zero code (%v): %v", keyString, resp.Code, resp.Log)
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

func accountCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return fmt.Errorf("account command requires an argument ([address])") //never stack trace
	}

	addrHex := StripHex(args[0])

	// convert destination address to bytes
	addr, err := hex.DecodeString(addrHex)
	if err != nil {
		return errors.Errorf("Account address (%v) is invalid hex: %v\n", addrHex, err)
	}

	acc, err := getAcc(nodeFlag, addr)
	if err != nil {
		return err
	}
	fmt.Println(string(wire.JSONBytes(acc)))
	return nil
}

func blockCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return fmt.Errorf("block command requires an argument ([height])") //never stack trace
	}

	heightString := args[0]
	height, err := strconv.Atoi(heightString)
	if err != nil {
		return errors.Errorf("Height must be an int, got %v: %v\n", heightString, err)
	}

	header, commit, err := getHeaderAndCommit(nodeFlag, height)
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

func verifyCmd(cmd *cobra.Command, args []string) error {

	keyString, valueString := keyFlag, valueFlag

	var err error
	key := []byte(keyString)
	if isHex(keyString) {
		key, err = hex.DecodeString(StripHex(keyString))
		if err != nil {
			return errors.Errorf("Key (%v) is invalid hex: %v\n", keyString, err)
		}
	}

	value := []byte(valueString)
	if isHex(valueString) {
		value, err = hex.DecodeString(StripHex(valueString))
		if err != nil {
			return errors.Errorf("Value (%v) is invalid hex: %v\n", valueString, err)
		}
	}

	root, err := hex.DecodeString(StripHex(rootFlag))
	if err != nil {
		return errors.Errorf("Root (%v) is invalid hex: %v\n", rootFlag, err)
	}

	proofBytes, err := hex.DecodeString(StripHex(proofFlag))
	if err != nil {
		return errors.Errorf("Proof (%v) is invalid hex: %v\n", proofBytes, err)
	}

	proof, err := merkle.ReadProof(proofBytes)
	if err != nil {
		return errors.Errorf("Error unmarshalling proof: %v\n", err)
	}

	if proof.Verify(key, value, root) {
		fmt.Println("OK")
	} else {
		return errors.New("Proof does not verify")
	}
	return nil
}
