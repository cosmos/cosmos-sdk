package keys

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"fmt"
	"os/signal"
	"os"
	"syscall"
	"bytes"
	"github.com/cosmos/cosmos-sdk/client"
	"bufio"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bip39"
)

const (
	flagEntropy = "user"
)

func mnemonicCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mnemonic",
		Short: "Creates a new mnemonic for use in key generation. Uses system entropy by default.",
		RunE:  runMnemonicCmd,
	}
	cmd.Flags().Bool(flagEntropy, false, "Prompt the use to enter entropy. Otherwise, use the system's entropy.")
	return cmd
}

func runMnemonicCmd(cmd *cobra.Command, args []string) error {
	kb, err := GetKeyBase()
	if err != nil {
		return err
	}

	if !viper.GetBool(flagEntropy) {
		return outputMnemonic(kb, nil)
	}

	stdin := client.BufferStdin()
	var buf bytes.Buffer
	done := make(chan bool)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	// need below signal handling in order to prevent panics on SIGTERM
	go func() {
		<-sigs
		fmt.Println("Killed.")
		os.Exit(1)
	}()

	go func() {
		fmt.Println("Please provide entropy using your keyboard and press enter.")
		scanner := bufio.NewScanner(stdin)
		for scanner.Scan() {
			buf.Write(scanner.Bytes())
			if buf.Len() >= bip39.FreshKeyEntropySize {
				done <- true
				return
			}

			fmt.Println("Please provide additional entropy and press enter.")
		}
	}()

	<-done
	if err != nil {
		return err
	}

	buf.Truncate(bip39.FreshKeyEntropySize)
	return outputMnemonic(kb, buf.Bytes())

}

func outputMnemonic(kb keys.Keybase, entropy []byte) error {
	fmt.Println("Generating mnemonic...")
	mnemonic, err := kb.GenerateMnemonic(keys.English, entropy)
	if err != nil {
		return err
	}

	fmt.Println(mnemonic)
	return nil
}