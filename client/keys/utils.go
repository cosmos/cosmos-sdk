package keys

import (
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	cryptokeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// available output formats.
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
)

type bechKeyOutFn func(k *cryptokeyring.Record) (cryptokeyring.KeyOutput, error)

func printKeyringRecord(w io.Writer, k *cryptokeyring.Record, bechKeyOut bechKeyOutFn, output string) {
	ko, err := bechKeyOut(k)
	if err != nil {
		panic(err)
	}

	switch output {
	case OutputFormatText:
		printTextRecords(w, []cryptokeyring.KeyOutput{ko})

	case OutputFormatJSON:
		out, err := KeysCdc.MarshalJSON(ko)
		if err != nil {
			panic(err)
		}

		fmt.Fprintln(w, string(out))
	}
}

func printKeyringRecords(w io.Writer, records []*cryptokeyring.Record, output string) {
	kos, err := cryptokeyring.MkAccKeysOutput(records)
	if err != nil {
		panic(err)
	}

	switch output {
	case OutputFormatText:
		printTextRecords(w, kos)

	case OutputFormatJSON:
		// TODO https://github.com/cosmos/cosmos-sdk/issues/8046
		// Replace AminoCdc with Proto JSON
		out, err := KeysCdc.MarshalJSON(kos)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "%s", out)
	}
}

func printTextRecords(w io.Writer, kos []cryptokeyring.KeyOutput) {
	out, err := yaml.Marshal(&kos)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(w, string(out))
}
