package stateviewer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/creachadair/tomledit"
)

// ReadOnlyDB is an interface for a database that can only be queried.
// We do not want user of StateViewer to be able to modify the database.
type ReadOnlyDB interface {
	// Get fetches the value of the given key, or nil if it does not exist.
	// CONTRACT: key, value readonly []byte
	Get([]byte) ([]byte, error)

	// Has checks if a key exists.
	// CONTRACT: key, value readonly []byte
	Has(key []byte) (bool, error)

	// Iterator returns an iterator over a domain of keys, in ascending order. The caller must call
	// Close when done. End is exclusive, and start must be less than end. A nil start iterates
	// from the first key, and a nil end iterates to the last key (inclusive). Empty keys are not
	// valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	Iterator(start, end []byte) (dbm.Iterator, error)

	// ReverseIterator returns an iterator over a domain of keys, in descending order. The caller
	// must call Close when done. End is exclusive, and start must be less than end. A nil end
	// iterates from the last key (inclusive), and a nil start iterates to the first key (inclusive).
	// Empty keys are not valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	ReverseIterator(start, end []byte) (dbm.Iterator, error)

	// Close closes the database connection.
	Close() error

	// Print is used for debugging.
	Print() error

	// Stats returns a map of property values for all keys and the size of the cache.
	Stats() map[string]string
}

type ReadDBConfig struct {
	BackendType dbm.BackendType
}

type ReadDBOption func(*ReadDBConfig)

func ReadDBOptionWithBackend(backendType string) ReadDBOption {
	return func(cfg *ReadDBConfig) {
		cfg.BackendType = dbm.BackendType(backendType)
	}
}

// ReadDB opens a database for reading.
// It reads from the application config file to determine the database backend, unless overridden.
func ReadDB(location string, options ...ReadDBOption) (ReadOnlyDB, error) {
	defaultDBBackend, err := getDBBackendFromConfig(location)
	if err != nil {
		return nil, err
	}

	config := &ReadDBConfig{
		BackendType: defaultDBBackend,
	}
	for _, opt := range options {
		opt(config)
	}

	// we do not need to check the backend type validaty as this will be done by the dbm package
	fmt.Printf("opening %s (backend) database\n", config.BackendType)
	db, err := openDB(location, config.BackendType)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func openDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}

func getDBBackendFromConfig(rootDir string) (dbm.BackendType, error) {
	path := filepath.Join(rootDir, "config")

	var configKeys = []struct {
		filename string
		key      string
	}{
		{"app.toml", "app-db-backend"},
		{"config.toml", "db_backend"},
	}

	var backend string
	for _, config := range configKeys {
		f, err := os.Open(filepath.Join(path, config.filename))
		if err != nil {
			return "", fmt.Errorf("failed to open %q: %w", config.filename, err)
		}
		defer f.Close()

		doc, err := tomledit.Parse(f)
		if err != nil {
			return "", fmt.Errorf("failed to parse %q: %w", config.filename, err)
		}

		results := doc.Find([]string{config.key}...)
		if len(results) == 0 {
			return "", fmt.Errorf("key %q not found", config.key)
		} else if len(results) > 1 {
			return "", fmt.Errorf("key %q is ambiguous", config.key)
		}

		backend = strings.ReplaceAll(results[0].Value.String(), `"`, "")
		if backend != "" {
			break
		}
	}

	return dbm.BackendType(backend), nil
}
