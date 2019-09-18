package keys

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/input"
)

func signCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign <name> <filename>",
		Short: "Sign a plain text payload with a private key and print the signed document to STDOUT",
		Long: `Sign an arbitrary text file with a private key and produce an amino-encoded JSON output.
The signed JSON document could eventually be verified through the 'keys verify' command and will
have the following structure:
{
  "text": original text file contents,
  "pub": public key,
  "sig": signature
}
`,
		Args: cobra.ExactArgs(2),
		RunE: runSignCmd,
	}
	cmd.SetOut(os.Stdout)
	return cmd
}

func runSignCmd(cmd *cobra.Command, args []string) error {
	name := args[0]
	filename := args[1]

	kb, err := NewKeyBaseFromHomeFlag()
	if err != nil {
		return err
	}

	msg, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	buf := bufio.NewReader(cmd.InOrStdin())
	passphrase, err := input.GetPassword(fmt.Sprintf("Password to sign with '%s':", name), buf)
	if err != nil {
		return err
	}

	sig, pub, err := kb.Sign(name, passphrase, msg)
	if err != nil {
		return err
	}

	out, err := MarshalJSON(signedText{
		Text: string(msg),
		Pub:  pub,
		Sig:  sig,
	})
	if err != nil {
		return err
	}

	cmd.Println(string(out))
	return nil
}
