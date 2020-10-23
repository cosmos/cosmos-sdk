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

func printKeyInfo(w io.Writer, keyInfo cryptokeyring.Info, bechKeyOut bechKeyOutFn, output string) error {
	ko, err := bechKeyOut(keyInfo)
	if err != nil {
		return err
	}

	switch output {
	case OutputFormatText:
		err = printTextInfos(w, []cryptokeyring.KeyOutput{ko})

	case OutputFormatJSON:
		out, err2 := KeysCdc.MarshalJSON(ko)
		if err2 != nil {
			return err2
		}
		_, err = fmt.Fprintln(w, string(out))
	}
	return err
}

func printInfos(w io.Writer, infos []cryptokeyring.Info, output string) error {
	kos, err := cryptokeyring.Bech32KeysOutput(infos)
	if err != nil {
		return err
	}

	switch output {
	case OutputFormatText:
		printTextInfos(w, kos)

	case OutputFormatJSON:
		out, err := KeysCdc.MarshalJSON(kos)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s", out); err != nil {
			return err
		}
	}
	return err
}

func printTextInfos(w io.Writer, kos []cryptokeyring.KeyOutput) error {
	out, err := yaml.Marshal(&kos)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(out))
	return err
}

func printKeyAddress(w io.Writer, info cryptokeyring.Info, bechKeyOut bechKeyOutFn) error {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprintln(w, ko.Address)
	return err
}

func printPubKey(w io.Writer, info cryptokeyring.Info, bechKeyOut bechKeyOutFn) error {
	ko, err := bechKeyOut(info)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, ko.Address)
	return err
}
