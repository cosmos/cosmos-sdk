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
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// available output formats.
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"

	// defaultKeyDBName is the client's subdirectory where keys are stored.
	defaultKeyDBName = "keys"
)

type bechKeyOutFn func(config *sdk.Config, keyInfo keys.Info) (keys.KeyOutput, error)

// NewKeyBaseFromHomeFlag initializes a Keybase based on the configuration. Keybase
// options can be applied when generating this new Keybase.
func NewKeyBaseFromHomeFlag(config *sdk.Config, opts ...keys.KeybaseOption) (keys.Keybase, error) {
	rootDir := viper.GetString(flags.FlagHome)
	return NewKeyBaseFromDir(rootDir, config, opts...)
}

// NewKeyBaseFromDir initializes a keybase at the rootDir directory. Keybase
// options can be applied when generating this new Keybase.
func NewKeyBaseFromDir(rootDir string, config *sdk.Config, opts ...keys.KeybaseOption) (keys.Keybase, error) {
	return keys.New(defaultKeyDBName, filepath.Join(rootDir, "keys"), config, opts...), nil
}

// NewInMemoryKeyBase returns a storage-less keybase.
func NewInMemoryKeyBase(config *sdk.Config) keys.Keybase { return keys.NewInMemory(config) }

// NewKeyBaseFromHomeFlag initializes a keyring based on configuration. Keybase
// options can be applied when generating this new Keybase.
func NewKeyringFromHomeFlag(input io.Reader, config *sdk.Config, opts ...keys.KeybaseOption) (keys.Keybase, error) {
	return NewKeyringFromDir(viper.GetString(flags.FlagHome), input, config, opts...)
}

// NewKeyBaseFromDir initializes a keyring at the given directory.
// If the viper flag flags.FlagKeyringBackend is set to file, it returns an on-disk keyring with
// CLI prompt support only. If flags.FlagKeyringBackend is set to test it will return an on-disk,
// password-less keyring that could be used for testing purposes.
func NewKeyringFromDir(rootDir string, input io.Reader, config *sdk.Config, opts ...keys.KeybaseOption) (keys.Keybase, error) {
	keyringBackend := viper.GetString(flags.FlagKeyringBackend)
	keyringName := config.GetKeyringServiceName()
	switch keyringBackend {
	case flags.KeyringBackendTest:
		return keys.NewTestKeyring(keyringName, rootDir, opts...)
	case flags.KeyringBackendFile:
		return keys.NewKeyringFile(keyringName, rootDir, input, opts...)
	case flags.KeyringBackendOS:
		return keys.NewKeyring(keyringName, rootDir, input, opts...)
	}
	return nil, fmt.Errorf("unknown keyring backend %q", keyringBackend)
}

func printKeyInfo(config *sdk.Config, keyInfo keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(config, keyInfo)
	if err != nil {
		panic(err)
	}

	switch viper.Get(cli.OutputFlag) {
	case OutputFormatText:
		printTextInfos([]keys.KeyOutput{ko})

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

		fmt.Println(string(out))
	}
}

func printInfos(config *sdk.Config, infos []keys.Info) {
	kos, err := keys.Bech32KeysOutput(config, infos)
	if err != nil {
		panic(err)
	}

	switch viper.Get(cli.OutputFlag) {
	case OutputFormatText:
		printTextInfos(kos)

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
		fmt.Printf("%s", out)
	}
}

func printTextInfos(kos []keys.KeyOutput) {
	out, err := yaml.Marshal(&kos)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

func printKeyAddress(config *sdk.Config, info keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(config, info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.Address)
}

func printPubKey(config *sdk.Config, info keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(config, info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.PubKey)
}

func isRunningUnattended() bool {
	backends := keyring.AvailableBackends()
	return len(backends) == 2 && backends[1] == keyring.BackendType("file")
}
