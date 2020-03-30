package keys

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/99designs/keyring"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptokeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// available output formats.
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"

	// defaultKeyDBName is the client's subdirectory where keys are stored.
	defaultKeyDBName = "keys"
)

type bechKeyOutFn func(keyInfo cryptokeyring.Info) (cryptokeyring.KeyOutput, error)

// NewKeyBaseFromDir initializes a keybase at the rootDir directory. Keybase
// options can be applied when generating this new Keybase.
func NewKeyBaseFromDir(rootDir string, opts ...cryptokeyring.KeybaseOption) (cryptokeyring.Keybase, error) {
	return getLazyKeyBaseFromDir(rootDir, opts...)
}

// NewInMemoryKeyBase returns a storage-less keybase.
func NewInMemoryKeyBase() cryptokeyring.Keybase { return cryptokeyring.NewInMemory() }

func getLazyKeyBaseFromDir(rootDir string, opts ...cryptokeyring.KeybaseOption) (cryptokeyring.Keybase, error) {
	return cryptokeyring.New(defaultKeyDBName, filepath.Join(rootDir, "keys"), opts...), nil
}

func printKeyInfo(w io.Writer, keyInfo cryptokeyring.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(keyInfo)
	if err != nil {
		panic(err)
	}

	switch viper.Get(cli.OutputFlag) {
	case OutputFormatText:
		printTextInfos(w, []cryptokeyring.KeyOutput{ko})

	case OutputFormatJSON:
		var out []byte
		var err error
		if viper.GetBool(flags.FlagIndentResponse) {
			out, err = KeysCdc.MarshalJSONIndent(ko, "", "  ")
		} else {
			out, err = KeysCdc.MarshalJSON(ko)
		}
		if err != nil {
			panic(err)
		}

		fmt.Fprintln(w, string(out))
	}
}

func printInfos(w io.Writer, infos []cryptokeyring.Info) {
	kos, err := cryptokeyring.Bech32KeysOutput(infos)
	if err != nil {
		panic(err)
	}

	switch viper.Get(cli.OutputFlag) {
	case OutputFormatText:
		printTextInfos(w, kos)

	case OutputFormatJSON:
		var out []byte
		var err error

		if viper.GetBool(flags.FlagIndentResponse) {
			out, err = KeysCdc.MarshalJSONIndent(kos, "", "  ")
		} else {
			out, err = KeysCdc.MarshalJSON(kos)
		}

		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "%s", out)
	}
}

func printTextInfos(w io.Writer, kos []cryptokeyring.KeyOutput) {
	out, err := yaml.Marshal(&kos)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(out))
}

func printKeyAddress(w io.Writer, info cryptokeyring.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(w, ko.Address)
}

func printPubKey(w io.Writer, info cryptokeyring.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(w, ko.PubKey)
}

func isRunningUnattended() bool {
	backends := keyring.AvailableBackends()
	return len(backends) == 2 && backends[1] == keyring.BackendType("file")
}
