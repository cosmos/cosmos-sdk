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

type bechKeyOutFn func(keyInfo keys.Info) (keys.KeyOutput, error)

// NewKeyBaseFromHomeFlag initializes a Keybase based on the configuration. Keybase
// options can be applied when generating this new Keybase.
func NewKeyBaseFromHomeFlag(opts ...keys.KeybaseOption) (keys.Keybase, error) {
	rootDir := viper.GetString(flags.FlagHome)
	return NewKeyBaseFromDir(rootDir, opts...)
}

// NewKeyBaseFromDir initializes a keybase at the rootDir directory. Keybase
// options can be applied when generating this new Keybase.
func NewKeyBaseFromDir(rootDir string, opts ...keys.KeybaseOption) (keys.Keybase, error) {
	return getLazyKeyBaseFromDir(rootDir, opts...)
}

// NewInMemoryKeyBase returns a storage-less keybase.
func NewInMemoryKeyBase() keys.Keybase { return keys.NewInMemory() }

// NewKeyBaseFromHomeFlag initializes a keyring based on configuration. Keybase
// options can be applied when generating this new Keybase.
func NewKeyringFromHomeFlag(input io.Reader, opts ...keys.KeybaseOption) (keys.Keybase, error) {
	return NewKeyringFromDir(viper.GetString(flags.FlagHome), input, opts...)
}

// NewKeyBaseFromDir initializes a keyring at the given directory.
// If the viper flag flags.FlagKeyringBackend is set to file, it returns an on-disk keyring with
// CLI prompt support only. If flags.FlagKeyringBackend is set to test it will return an on-disk,
// password-less keyring that could be used for testing purposes.
func NewKeyringFromDir(rootDir string, input io.Reader, opts ...keys.KeybaseOption) (keys.Keybase, error) {
	keyringBackend := viper.GetString(flags.FlagKeyringBackend)
	switch keyringBackend {
	case flags.KeyringBackendTest:
		return keys.NewTestKeyring(sdk.GetConfig().GetKeyringServiceName(), rootDir, opts...)
	case flags.KeyringBackendFile:
		return keys.NewKeyringFile(sdk.GetConfig().GetKeyringServiceName(), rootDir, input, opts...)
	case flags.KeyringBackendOS:
		return keys.NewKeyring(sdk.GetConfig().GetKeyringServiceName(), rootDir, input, opts...)
	}
	return nil, fmt.Errorf("unknown keyring backend %q", keyringBackend)
}

func getLazyKeyBaseFromDir(rootDir string, opts ...keys.KeybaseOption) (keys.Keybase, error) {
	return keys.New(defaultKeyDBName, filepath.Join(rootDir, "keys"), opts...), nil
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

func isRunningUnattended() bool {
	backends := keyring.AvailableBackends()
	return len(backends) == 2 && backends[1] == keyring.BackendType("file")
}
