package commands

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-merkle"
	"github.com/tendermint/go-wire"
	tmtypes "github.com/tendermint/tendermint/types"
)

//commands
var (
	QueryCmd = &cobra.Command{
		Use:   "query [key]",
		Short: "Query the merkle tree",
		Run:   queryCmd,
	}

	AccountCmd = &cobra.Command{
		Use:   "account [address]",
		Short: "Get details of an account",
		Run:   accountCmd,
	}

	BlockCmd = &cobra.Command{
		Use:   "block [height]",
		Short: "Get the header and commit of a block",
		Run:   blockCmd,
	}

	VerifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "Verify the IAVL proof",
		Run:   verifyCmd,
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

func queryCmd(cmd *cobra.Command, args []string) {

	if len(args) != 1 {
		cmn.Exit("query command requires an argument ([key])")
	}

	keyString := args[0]
	key := []byte(keyString)
	if isHex(keyString) {
		// convert key to bytes
		var err error
		key, err = hex.DecodeString(StripHex(keyString))
		if err != nil {
			cmn.Exit(fmt.Sprintf("Query key (%v) is invalid hex: %+v\n", keyString, err))
		}
	}

	resp, err := Query(nodeFlag, key)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Query returns error: %+v\n", err))
	}

	if !resp.Code.IsOK() {
		cmn.Exit(fmt.Sprintf("Query for key (%v) returned non-zero code (%v): %v", keyString, resp.Code, resp.Log))
	}

	val := resp.Value
	proof := resp.Proof
	height := resp.Height

	fmt.Println(string(wire.JSONBytes(struct {
		Value  []byte `json:"value"`
		Proof  []byte `json:"proof"`
		Height uint64 `json:"height"`
	}{val, proof, height})))
}

func accountCmd(cmd *cobra.Command, args []string) {

	if len(args) != 1 {
		cmn.Exit("account command requires an argument ([address])")
	}

	addrHex := StripHex(args[0])

	// convert destination address to bytes
	addr, err := hex.DecodeString(addrHex)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Account address (%v) is invalid hex: %+v\n", addrHex, err))
	}

	acc, err := getAcc(nodeFlag, addr)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}
	fmt.Println(string(wire.JSONBytes(acc)))
}

func blockCmd(cmd *cobra.Command, args []string) {

	if len(args) != 1 {
		cmn.Exit("block command requires an argument ([height])")
	}

	heightString := args[0]
	height, err := strconv.Atoi(heightString)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Height must be an int, got %v: %+v\n", heightString, err))
	}

	header, commit, err := getHeaderAndCommit(nodeFlag, height)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
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
}

type BlockHex struct {
	Header []byte `json:"header"`
	Commit []byte `json:"commit"`
}

type BlockJSON struct {
	Header *tmtypes.Header `json:"header"`
	Commit *tmtypes.Commit `json:"commit"`
}

func verifyCmd(cmd *cobra.Command, args []string) {

	keyString, valueString := keyFlag, valueFlag

	var err error
	key := []byte(keyString)
	if isHex(keyString) {
		key, err = hex.DecodeString(StripHex(keyString))
		if err != nil {
			cmn.Exit(fmt.Sprintf("Key (%v) is invalid hex: %+v\n", keyString, err))
		}
	}

	value := []byte(valueString)
	if isHex(valueString) {
		value, err = hex.DecodeString(StripHex(valueString))
		if err != nil {
			cmn.Exit(fmt.Sprintf("Value (%v) is invalid hex: %+v\n", valueString, err))
		}
	}

	root, err := hex.DecodeString(StripHex(rootFlag))
	if err != nil {
		cmn.Exit(fmt.Sprintf("Root (%v) is invalid hex: %+v\n", rootFlag, err))
	}

	proofBytes, err := hex.DecodeString(StripHex(proofFlag))

	proof, err := merkle.ReadProof(proofBytes)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error unmarshalling proof: %+v\n", err))
	}

	if proof.Verify(key, value, root) {
		fmt.Println("OK")
	} else {
		cmn.Exit(fmt.Sprintf("Proof does not verify"))
	}
}
