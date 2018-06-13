package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	gaia "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/spf13/cobra"
	crypto "github.com/tendermint/go-crypto"
)

func init() {
	rootCmd.AddCommand(txCmd)
	rootCmd.AddCommand(pubkeyCmd)
	rootCmd.AddCommand(hackCmd)
	rootCmd.AddCommand(rawBytesCmd)
}

var rootCmd = &cobra.Command{
	Use:          "gaiadebug",
	Short:        "Gaia debug tool",
	SilenceUsage: true,
}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "Decode a gaia tx from hex or base64",
	RunE:  runTxCmd,
}

var pubkeyCmd = &cobra.Command{
	Use:   "pubkey",
	Short: "Decode a pubkey from hex or base64",
	RunE:  runPubKeyCmd,
}

var hackCmd = &cobra.Command{
	Use:   "hack",
	Short: "Boilerplate to Hack on an existing state by scripting some Go...",
	RunE:  runHackCmd,
}

var rawBytesCmd = &cobra.Command{
	Use:   "raw-bytes",
	Short: "Convert raw bytes output (eg. [10 21 13 255]) to hex",
	RunE:  runRawBytesCmd,
}

func runRawBytesCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Expected single arg")
	}
	stringBytes := args[0]
	stringBytes = strings.Trim(stringBytes, "[")
	stringBytes = strings.Trim(stringBytes, "]")
	spl := strings.Split(stringBytes, " ")

	byteArray := []byte{}
	for _, s := range spl {
		b, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		byteArray = append(byteArray, byte(b))
	}
	fmt.Printf("%X\n", byteArray)
	return nil
}

func runPubKeyCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Expected single arg")
	}

	pubkeyString := args[0]

	// try hex, then base64
	pubkeyBytes, err := hex.DecodeString(pubkeyString)
	if err != nil {
		var err2 error
		pubkeyBytes, err2 = base64.StdEncoding.DecodeString(pubkeyString)
		if err2 != nil {
			return fmt.Errorf(`Expected hex or base64. Got errors:
			hex: %v,
			base64: %v
			`, err, err2)
		}
	}

	cdc := gaia.MakeCodec()
	var pubKey crypto.PubKeyEd25519
	copy(pubKey[:], pubkeyBytes)
	pubKeyJSONBytes, err := cdc.MarshalJSON(pubKey)
	if err != nil {
		return err
	}
	fmt.Println("Address:", pubKey.Address())
	fmt.Printf("Hex: %X\n", pubkeyBytes)
	fmt.Println("JSON (base64):", string(pubKeyJSONBytes))
	return nil
}

func runTxCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Expected single arg")
	}

	txString := args[0]

	// try hex, then base64
	txBytes, err := hex.DecodeString(txString)
	if err != nil {
		var err2 error
		txBytes, err2 = base64.StdEncoding.DecodeString(txString)
		if err2 != nil {
			return fmt.Errorf(`Expected hex or base64. Got errors:
			hex: %v,
			base64: %v
			`, err, err2)
		}
	}

	var tx = auth.StdTx{}
	cdc := gaia.MakeCodec()

	err = cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return err
	}

	bz, err := cdc.MarshalJSON(tx)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte{})
	err = json.Indent(buf, bz, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(buf.String())
	return nil
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
