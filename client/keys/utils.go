package keys

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

// available output formats.
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"

	// defaultKeyDBName is the client's subdirectory where keys are stored.
	defaultKeyDBName = "keys"
)

type bechKeyOutFn func(keyInfo keys.Info) (keys.KeyOutput, error)

// GetKeyInfo returns key info for a given name. An error is returned if the
// keybase cannot be retrieved or getting the info fails.
func GetKeyInfo(name string) (keys.Info, error) {
	keybase, err := NewKeyBaseFromHomeFlag()
	if err != nil {
		return nil, err
	}

	return keybase.Get(name)
}

// GetPassphrase returns a passphrase for a given name. It will first retrieve
// the key info for that name if the type is local, it'll fetch input from
// STDIN. Otherwise, an empty passphrase is returned. An error is returned if
// the key info cannot be fetched or reading from STDIN fails.
func GetPassphrase(name string) (string, error) {
	var passphrase string

	keyInfo, err := GetKeyInfo(name)
	if err != nil {
		return passphrase, err
	}

	// we only need a passphrase for locally stored keys
	// TODO: (ref: #864) address security concerns
	if keyInfo.GetType() == keys.TypeLocal {
		passphrase, err = ReadPassphraseFromStdin(name)
		if err != nil {
			return passphrase, err
		}
	}

	return passphrase, nil
}

// ReadPassphraseFromStdin attempts to read a passphrase from STDIN return an
// error upon failure.
func ReadPassphraseFromStdin(name string) (string, error) {
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)

	passphrase, err := client.GetPassword(prompt, buf)
	if err != nil {
		return passphrase, fmt.Errorf("Error reading passphrase: %v", err)
	}

	return passphrase, nil
}

// NewKeyBaseFromHomeFlag initializes a Keybase based on the configuration.
func NewKeyBaseFromHomeFlag() (keys.Keybase, error) {
	rootDir := viper.GetString(cli.HomeFlag)
	return NewKeyBaseFromDir(rootDir)
}

// NewKeyBaseFromDir initializes a keybase at a particular dir.
func NewKeyBaseFromDir(rootDir string) (keys.Keybase, error) {
	return getLazyKeyBaseFromDir(rootDir)
}

// NewInMemoryKeyBase returns a storage-less keybase.
func NewInMemoryKeyBase() keys.Keybase { return keys.NewInMemory() }

func getLazyKeyBaseFromDir(rootDir string) (keys.Keybase, error) {
	return keys.New(defaultKeyDBName, filepath.Join(rootDir, "keys")), nil
}

func printMultiSigKeyInfo(keyInfo keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(keyInfo)
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"WEIGHT", "THRESHOLD", "ADDRESS", "PUBKEY"})
	threshold := fmt.Sprintf("%d", ko.Threshold)
	for _, pk := range ko.PubKeys {
		weight := fmt.Sprintf("%d", pk.Weight)
		table.Append([]string{weight, threshold, pk.Address, pk.PubKey})
	}
	table.Render()
}

func printKeyInfo(keyInfo keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(keyInfo)
	if err != nil {
		panic(err)
	}

	switch viper.Get(cli.OutputFlag) {
	case OutputFormatText:
		printTextInfos([]keys.KeyOutput{ko})

	case OutputFormatJSON:
		var out []byte
		var err error
		if viper.GetBool(client.FlagIndentResponse) {
			out, err = cdc.MarshalJSONIndent(ko, "", "  ")
		} else {
			out, err = cdc.MarshalJSON(ko)
		}
		if err != nil {
			panic(err)
		}

		fmt.Println(string(out))
	}
}

func printInfos(infos []keys.Info) {
	kos, err := keys.Bech32KeysOutput(infos)
	if err != nil {
		panic(err)
	}

	switch viper.Get(cli.OutputFlag) {
	case OutputFormatText:
		printTextInfos(kos)

	case OutputFormatJSON:
		var out []byte
		var err error

		if viper.GetBool(client.FlagIndentResponse) {
			out, err = cdc.MarshalJSONIndent(kos, "", "  ")
		} else {
			out, err = cdc.MarshalJSON(kos)
		}

		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	}
}

func printTextInfos(kos []keys.KeyOutput) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "TYPE", "ADDRESS", "PUBKEY"})
	for _, ko := range kos {
		table.Append([]string{ko.Name, ko.Type, ko.Address, ko.PubKey})
	}
	table.Render()
}

func printKeyAddress(info keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.Address)
}

func printPubKey(info keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.PubKey)
}
