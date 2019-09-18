package keys

import (
	"errors"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func verifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify <filename>",
		Short: "Verify a plain text's signature",
		Long: `Read a document generated with the 'key sign' command and verify the signature.
It exits with 0 if the signature verification succeed; it returns a value different than 0
if the signature verification fails.
`,
		Args: cobra.ExactArgs(1),
		RunE: runVerifyCmd,
	}
	return cmd
}

func runVerifyCmd(cmd *cobra.Command, args []string) error {
	filename := args[0]

	signedDoc, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var doc signedText
	if err := UnmarshalJSON(signedDoc, &doc); err != nil {
		return err
	}

	if doc.Pub.VerifyBytes([]byte(doc.Text), doc.Sig) {
		cmd.PrintErrln("signature verified")
		return nil
	}
	return errors.New("bad signature")
}
