package cosmovisor

import (
	"errors"
	"fmt"
	"strings"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/pelletier/go-toml/v2"
)

// parseDBBackend extracts the db_backend value from the output of
// `<daemon> config get config db_backend`. Newer SDK versions print the value
// TOML-encoded (e.g. "goleveldb" with quotes and a trailing newline), older
// ones may print it bare. The result is only syntactically validated:
// dbm.NewDB remains responsible for rejecting unknown or unbuilt backends.
func parseDBBackend(out []byte) (dbm.BackendType, error) {
	value := strings.Trim(string(out), " \t\r\n")
	if value == "" {
		return "", errors.New("empty db_backend value")
	}
	if strings.ContainsAny(value, "\r\n") {
		return "", errors.New("db_backend output spans multiple lines")
	}

	if q := value[0]; q == '"' || q == '\'' {
		if len(value) < 2 || value[len(value)-1] != q {
			return "", errors.New("unterminated quoted db_backend value")
		}
		var doc struct {
			V string `toml:"v"`
		}
		if err := toml.Unmarshal([]byte("v = "+value), &doc); err != nil {
			return "", fmt.Errorf("decode quoted db_backend value: %w", err)
		}
		value = doc.V
	}

	if value == "" {
		return "", errors.New("empty db_backend value")
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return "", fmt.Errorf("invalid character %q in db_backend value", r)
		}
	}
	return dbm.BackendType(value), nil
}
