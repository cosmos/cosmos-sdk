package keys

import (
	"fmt"
	"io"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

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

// NewLegacyKeyBaseFromDir initializes a legacy keybase at the rootDir directory. Keybase
// options can be applied when generating this new Keybase.
func NewLegacyKeyBaseFromDir(rootDir string, opts ...cryptokeyring.KeybaseOption) (cryptokeyring.LegacyKeybase, error) {
	return getLegacyKeyBaseFromDir(rootDir, opts...)
}

func getLegacyKeyBaseFromDir(rootDir string, opts ...cryptokeyring.KeybaseOption) (cryptokeyring.LegacyKeybase, error) {
	return cryptokeyring.NewLegacy(defaultKeyDBName, filepath.Join(rootDir, "keys"), opts...)
}

func printKeyInfo(w io.Writer, keyInfo cryptokeyring.Info, bechKeyOut bechKeyOutFn, output string) {
	ko, err := bechKeyOut(keyInfo)
	if err != nil {
		panic(err)
	}

	switch output {
	case OutputFormatText:
		printTextInfos(w, []cryptokeyring.KeyOutput{ko})

	case OutputFormatJSON:
		out, err := KeysCdc.MarshalJSON(ko)
		if err != nil {
			panic(err)
		}

		fmt.Fprintln(w, string(out))
	}
}

func printInfos(w io.Writer, infos []cryptokeyring.Info, output string) {
	kos, err := cryptokeyring.MkAccKeysOutput(infos)
	if err != nil {
		panic(err)
	}

	switch output {
	case OutputFormatText:
		printTextInfos(w, kos)

	case OutputFormatJSON:
		out, err := KeysCdc.MarshalJSON(kos)
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
