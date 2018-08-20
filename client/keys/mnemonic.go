package keys

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	bip39 "github.com/bartekn/go-bip39"
)

const (
	flagUserEntropy = "user"
)

const (
	mnemonicEntropySize = 256
)

func mnemonicKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mnemonic",
		Short: "Compute the bip39 mnemonic for some input entropy",
		Long:  "Create a bip39 mnemonic, sometimes called a seed phrase, by reading from the system entropy. To pass your own entropy, use --user",
		RunE:  runMnemonicCmd,
	}
	cmd.Flags().Bool(flagUserEntropy, false, "Prompt the user to supply their own entropy, instead of relying on the system")
	return cmd
}

func runMnemonicCmd(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	userEntropy, _ := flags.GetBool(flagUserEntropy)

	var entropySeed []byte

	if userEntropy {
		// prompt the user to enter some entropy
		fmt.Println("> Roll 256-bits of dice entropy and enter the results here:")
		reader := bufio.NewReader(os.Stdin)
		inputEntropy, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		fmt.Println("> Input length:", len(inputEntropy))
		if len(inputEntropy) < 99 {
			fmt.Println("> WARNING! 256-bits of entropy is ~99 rolls of a 6-sided die")
		}

		// hash input entropy to get entropy seed
		hashedEntropy := sha256.Sum256([]byte(inputEntropy))
		entropySeed = hashedEntropy[:]
		fmt.Println("-------------------------------------------------------")
	} else {
		// read entropy seed straight from crypto.Rand
		var err error
		entropySeed, err = bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return err
		}
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		return err
	}

	fmt.Println(mnemonic)

	return nil
}
