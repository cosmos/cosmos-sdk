package cli

import (
	"fmt"
	"os"
	"encoding/json"
	"strings"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring" 
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
)

func GetMultiSignCommand(codec *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisign [in-file] [out-file]",
		Short: "Sign many standard transactions generated offline",
		Long: `Sign a list of transactions created with the --generate-only flag.
It will read StdSignDoc JSONs from [in-file], one transaction per line, and
produce a file of JSON encoded StdSignatures, one per line.

This command is intended to work offline for security purposes.


`,
		PreRun: preSignCmd,
		RunE:   makeMultiSignCmd(codec),
		Args:   cobra.ExactArgs(2),
	}

	cmd = flags.PostCommands(cmd)[0]
	cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}

func makeMultiSignCmd(cdc *codec.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		var fromName = viper.GetString(flags.FlagFrom)
		var inFileName = args[0]
		var outFileName = args[1]

		// Get TxBuilder from cli context
                var keybase, err = keyring.New(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), os.Stdin)
                if err != nil {
                        panic(err)
                } 

		// Parse StdSignDocs from file.
		var docs = parseStdSignDocsFromFile(inFileName)

		// Collect signatures.
		var sigs []types.StdSignature = nil
		for _, doc := range docs {
			sig, pubkey, err := keybase.Sign(fromName, doc.Bytes())                                                  
			if err != nil { panic(err) }
			sigs = append(sigs, types.StdSignature{
				PubKey:pubkey.Bytes(),
				Signature: sig,
			})

			// TODO: at this point, it is possible to derive the address from the pubkey,
			// and so we could auto-inject the signature in the appropriate place,
			// *if* we could also deal with multisig.  But I'm not sure how
			// best to solve this (ideally the cli flags + the keybase includes enough information
			// to figure it out based on the GetSigners(), but that isn't available
			// unless the sdk.Msg is parsed out.
			// Given that this function is intended to work only on StdSignDoc's w/
			// json.RawMessage for msgs, there just isn't enough info to do this automatically,
			// but perhaps there can be a "sig-position" flag.
			// Until then, or something else, --signature-only is not supported.
		}

		// Write StdSignatures to file.
		fp, err := os.OpenFile(
			outFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644,
		)
		if err != nil { panic(err) }
		defer fp.Close()

		for _, sig := range sigs {
			sigJSON, err := json.Marshal(sig)
			if err != nil { panic(err) }
			fmt.Fprintf(fp, "%X\n", sigJSON)
		}
		//fp.Flush()
		return nil
	}
}

//----------------------------------------

func parseStdSignDocsFromFile(filename string) (docs []types.StdSignDoc) {
	body, err := ioutil.ReadFile(filename)
	if err != nil { panic(err) }
	return parseStdSignDocs(string(body))
}

func parseStdSignDocs(lines string) (docs []types.StdSignDoc) {
	linez := strings.Split(lines, "\n")
	for _, line := range linez {
		docs = append(docs, parseStdSignDoc(line))
	}
	return docs
}

func parseStdSignDoc(line string) (doc types.StdSignDoc) {
	err := json.Unmarshal([]byte(line), &doc)
	if err != nil { panic(err) }
	// To validate, encode using standard encoding and make sure the line is the same.
	// If we mis-typed a field, or if there were duplicate fields, it will show here.
	if line != string(doc.Bytes()) {
		panic(fmt.Sprintf("%v is not valid StdSignDoc; expected %v",
			line, string(doc.Bytes())))
	}
	return
}
