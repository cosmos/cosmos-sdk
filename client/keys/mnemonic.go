package keys

import (
	"bufio"
	"crypto/sha256"
	"fmt"

	bip39 "github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/input"
)

const (
	flagUserEntropy = "unsafe-entropy"

	mnemonicEntropySize = 256
)

// MnemonicKeyCommand computes the bip39 memonic for input entropy.
func MnemonicKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mnemonic",
		Short: "Compute the bip39 mnemonic for some input entropy",
		Long:  "Create a bip39 mnemonic, sometimes called a seed phrase, by reading from the system entropy. To pass your own entropy, use --unsafe-entropy",
		RunE: func(cmd *cobra.Command, args []string) error {
			var entropySeed []byte

			if userEntropy, _ := cmd.Flags().GetBool(flagUserEntropy); userEntropy {
				// prompt the user to enter some entropy
				buf := bufio.NewReader(cmd.InOrStdin())

				inputEntropy, err := input.GetString("> WARNING: Generate at least 256-bits of entropy and enter the results here:", buf)
				if err != nil {
					return err
				}

				if len(inputEntropy) < 43 {
					return fmt.Errorf("256-bits is 43 characters in Base-64, and 100 in Base-6. You entered %v, and probably want more", len(inputEntropy))
				}

				conf, err := input.GetConfirmation(fmt.Sprintf("> Input length: %d", len(inputEntropy)), buf, cmd.ErrOrStderr())
				if err != nil {
					return err
				}

				if !conf {
					return nil
				}

				// hash input entropy to get entropy seed
				hashedEntropy := sha256.Sum256([]byte(inputEntropy))
				entropySeed = hashedEntropy[:]
			} else {
				// read entropy seed straight from crypto.Rand
				var err error
				entropySeed, err = bip39.NewEntropy(mnemonicEntropySize)
				if err != nil {
					return err
				}
			}

			mnemonic, err := bip39.NewMnemonic(entropySeed)
			if err != nil {
				return err
			}

			cmd.Println(mnemonic)
			return nil
		},
	}

	cmd.Flags().Bool(flagUserEntropy, false, "Prompt the user to supply their own entropy, instead of relying on the system")
	return cmd
}
