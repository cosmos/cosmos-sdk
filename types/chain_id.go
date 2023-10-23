package types

import (
	"encoding/json"
	"errors"
	fmt "fmt"
	"io"

	"github.com/cometbft/cometbft/types"
)

const ChainIDFieldName = "chain_id"

// ParseChainIDFromGenesis parses the `chain_id` from the genesis json file and abort early,
// it still parses the values before the `chain_id` field, particularly if the `app_state` field is
// before the `chain_id` field, it will parse the `app_state` value, user must make sure the `chain_id`
// is put before `app_state` or other big entries to enjoy the efficiency.
func ParseChainIDFromGenesis(r io.Reader) (string, error) {
	dec := json.NewDecoder(r)

	var t json.Token
	t, err := dec.Token()
	if err != nil {
		return "", err
	}
	if t != json.Delim('{') {
		return "", fmt.Errorf("expected {, got %s", t)
	}

	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return "", err
		}
		key, ok := t.(string)
		if !ok {
			return "", fmt.Errorf("expected string, got %s", t)
		}
		var value interface{}
		err = dec.Decode(&value)
		if err != nil {
			return "", err
		}
		if key == ChainIDFieldName {
			chainId, ok := value.(string)
			if !ok {
				return "", fmt.Errorf("expected string chain_id, got %s", value)
			}
			if err := validateChainID(chainId); err != nil {
				return "", err
			}
			return chainId, nil
		}
	}

	return "", errors.New("missing chain-id in genesis file")
}

func validateChainID(chainID string) error {
	if chainID == "" {
		return errors.New("genesis doc must include non-empty chain_id")
	}
	if len(chainID) > types.MaxChainIDLen {
		return fmt.Errorf("chain_id in genesis doc is too long (max: %d)", types.MaxChainIDLen)
	}

	return nil
}
