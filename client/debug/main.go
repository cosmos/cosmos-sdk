package debug

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

func Cmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Tool for helping with debugging your application",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(PubkeyCmd(cdc))
	cmd.AddCommand(AddrCmd())
	cmd.AddCommand(RawBytesCmd())

	return cmd
}

// getPubKeyFromString returns a Tendermint PubKey (PubKeyEd25519) by attempting
// to decode the pubkey string from hex, base64, and finally bech32. If all
// encodings fail, an error is returned.
func getPubKeyFromString(pkstr string) (crypto.PubKey, error) {
	var pubKey ed25519.PubKeyEd25519

	bz, err := hex.DecodeString(pkstr)
	if err == nil {
		copy(pubKey[:], bz)
		return pubKey, nil
	}

	bz, err = base64.StdEncoding.DecodeString(pkstr)
	if err == nil {
		copy(pubKey[:], bz)
		return pubKey, nil
	}

	pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, pkstr)
	if err == nil {
		return pk, nil
	}

	pk, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeValPub, pkstr)
	if err == nil {
		return pk, nil
	}

	pk, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, pkstr)
	if err == nil {
		return pk, nil
	}

	return nil, fmt.Errorf("pubkey '%s' invalid; expected hex, base64, or bech32", pubKey)
}

func PubkeyCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pubkey [pubkey]",
		Short: "Decode a ED25519 pubkey from hex, base64, or bech32",
		Long: fmt.Sprintf(`Decode a pubkey from hex, base64, or bech32.

Example:
$ %s debug pubkey TWFuIGlzIGRpc3Rpbmd1aXNoZWQsIG5vdCBvbmx5IGJ5IGhpcyByZWFzb24sIGJ1dCBieSB0aGlz
$ %s debug pubkey cosmos1e0jnq2sun3dzjh8p2xq95kk0expwmd7shwjpfg
			`, version.ClientName, version.ClientName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pk, err := getPubKeyFromString(args[0])
			if err != nil {
				return err
			}

			edPK, ok := pk.(ed25519.PubKeyEd25519)
			if !ok {
				return fmt.Errorf("invalid pubkey type; expected ED25519")
			}

			pubKeyJSONBytes, err := cdc.MarshalJSON(edPK)
			if err != nil {
				return err
			}
			accPub, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, edPK)
			if err != nil {
				return err
			}
			valPub, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeValPub, edPK)
			if err != nil {
				return err
			}
			consenusPub, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, edPK)
			if err != nil {
				return err
			}

			cmd.Println("Address:", edPK.Address())
			cmd.Printf("Hex: %X\n", edPK[:])
			cmd.Println("JSON (base64):", string(pubKeyJSONBytes))
			cmd.Println("Bech32 Acc:", accPub)
			cmd.Println("Bech32 Validator Operator:", valPub)
			cmd.Println("Bech32 Validator Consensus:", consenusPub)

			return nil
		},
	}
}

func AddrCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "addr [address]",
		Short: "Convert an address between hex and bech32",
		Long: fmt.Sprintf(`Convert an address between hex encoding and bech32.
			
Example:
$ %s debug addr cosmos1e0jnq2sun3dzjh8p2xq95kk0expwmd7shwjpfg
			`, version.ClientName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			addrString := args[0]
			var addr []byte

			// try hex, then bech32
			var err error
			addr, err = hex.DecodeString(addrString)
			if err != nil {
				var err2 error
				addr, err2 = sdk.AccAddressFromBech32(addrString)
				if err2 != nil {
					var err3 error
					addr, err3 = sdk.ValAddressFromBech32(addrString)

					if err3 != nil {
						return fmt.Errorf("expected hex or bech32. Got errors: hex: %v, bech32 acc: %v, bech32 val: %v", err, err2, err3)

					}
				}
			}

			accAddr := sdk.AccAddress(addr)
			valAddr := sdk.ValAddress(addr)

			cmd.Println("Address:", addr)
			cmd.Printf("Address (hex): %X\n", addr)
			cmd.Printf("Bech32 Acc: %s\n", accAddr)
			cmd.Printf("Bech32 Val: %s\n", valAddr)
			return nil
		},
	}
}

func RawBytesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "raw-bytes [raw-bytes]",
		Short: "Convert raw bytes output (eg. [10 21 13 255]) to hex",
		Long: fmt.Sprintf(`Convert raw-bytes to hex.
			
Example:
$ %s debug raw-bytes [72 101 108 108 111 44 32 112 108 97 121 103 114 111 117 110 100]
			`, version.ClientName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
}
