package keys

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/cli"
	"testing"
)

const (
	testMnemonic1 = "snake island check kitten mobile member toe asthma betray symptom adjust maximum"
)

func TestPrintCreateEmpty(t *testing.T) {
	err := printCreate(nil, false, "")
	assert.NotNil(t, err)
}

func ExamplePrintNewAddressText() {
	var kb keys.Keybase
	kb = client.MockKeyBase()

	info, err := kb.CreateAccount("some_name", testMnemonic1, defaultBIP39Passphrase, "", 0, 0)

	if err != nil {
		fmt.Println("ERROR")
	}

	viper.Set(cli.OutputFlag, "text")
	err = printCreate(info, true, "")

	// Output:
	//NAME:	TYPE:	ADDRESS:						PUBKEY:
	//some_name	offline	cosmos143cm6xgtpwjm2s0dmmh5538cse478gn4eafsu8	cosmospub1addwnpepq0n89mkx74tykj8kwk443v6jjwqqylgnx0xnwdwgk34uakw7666nqmj8a8j
	//
	//**Important** write this mnemonic phrase in a safe place.It is the only way to recover your account if you ever forget your password.
}

func ExamplePrintNewAddressJson() {
	var kb keys.Keybase
	kb = client.MockKeyBase()

	info, err := kb.CreateAccount("some_name", testMnemonic1, defaultBIP39Passphrase, "", 0, 0)

	if err != nil {
		fmt.Println("ERROR")
	}

	viper.Set(cli.OutputFlag, "json")
	err = printCreate(info, true, "")

	// Output:
	// {"name":"some_name","type":"offline","address":"cosmos143cm6xgtpwjm2s0dmmh5538cse478gn4eafsu8","pub_key":"cosmospub1addwnpepq0n89mkx74tykj8kwk443v6jjwqqylgnx0xnwdwgk34uakw7666nqmj8a8j"}
}

func ExampleCLIAddNewAddress() {
	cmd := cobra.Command{}
	args := []string{"test1"}

	viper.Set(flagDryRun, true)
	viper.Set(cli.OutputFlag, "text")
	viper.Set(flagRecover, true)

	err := runAddCmd(&cmd, args)
	if err != nil {
		fmt.Printf(err.Error())
	}

	// TODO: Enable mocking
	// Output:
	// EOF
}
