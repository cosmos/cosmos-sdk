package confix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/creachadair/atomicfile"
	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/transform"
	"github.com/spf13/viper"

	clientcfg "github.com/cosmos/cosmos-sdk/client/config"
	srvcfg "github.com/cosmos/cosmos-sdk/server/config"
)

// Migrate reads the configuration file at configPath and applies any
// transformations necessary to migrate it to the current version. If this
// succeeds, the transformed output is written to outputPath. As a special
// case, if outputPath == "" the output is written to stdout.
//
// It is safe if outputPath == inputPath. If a regular file outputPath already
// exists, it is overwritten. In case of error, the output is not written.
//
// Migrate is a convenience wrapper for calls to LoadConfig, ApplyFixes, and
// CheckValid. If the caller requires more control over the behavior of the
// migrate, call those functions directly.
func Migrate(ctx context.Context, sdkVersion, configPath, outputPath string) error {
	if configPath == "" {
		return errors.New("empty input configuration path")
	}

	doc, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %v", err)
	}

	if err := ApplyFixes(ctx, doc); err != nil {
		return fmt.Errorf("updating %q: %v", configPath, err)
	}

	var buf bytes.Buffer
	if err := tomledit.Format(&buf, doc); err != nil {
		return fmt.Errorf("formatting config: %v", err)
	}

	// verify that file is valid after applying fixes
	if err := CheckValid(configPath, buf.Bytes()); err != nil {
		return fmt.Errorf("updated config is invalid: %v", err)
	}

	if outputPath == "" {
		_, err = os.Stdout.Write(buf.Bytes())
	} else {
		err = atomicfile.WriteData(outputPath, buf.Bytes(), 0600)
	}

	return err
}

// ApplyFixes transforms doc and reports whether it succeeded.
func ApplyFixes(ctx context.Context, doc *tomledit.Document) error {
	return plan.Apply(ctx, doc)
}

// LoadConfig loads and parses the TOML document from path.
func LoadConfig(path string) (*tomledit.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return tomledit.Parse(f)
}

// CheckValid checks whether the specified config appears to be a valid Cosmos SDK config file.
// It tries to unmarshal the config into both the server and client config structs.
func CheckValid(fileName string, data []byte) error {
	v := viper.New()
	v.SetConfigType("toml")

	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	switch fileName {
	case AppConfig:
		var cfg srvcfg.Config
		if err := v.Unmarshal(&cfg); err != nil {
			return fmt.Errorf("failed to unmarshal as server config: %w", err)
		}

		return cfg.ValidateBasic()
	case ClientConfig:
		var cfg clientcfg.ClientConfig
		if err := v.Unmarshal(&cfg); err != nil {
			return fmt.Errorf("failed to unmarshal as config config: %w", err)
		}
	case TMConfig:
		return errors.New("tendermint config is not supported yet")
	default:
		return fmt.Errorf("unknown config file name: %s", fileName)
	}

	return nil
}

// WithLogWriter returns a child of ctx with a logger attached that sends
// output to w. This is a convenience wrapper for transform.WithLogWriter.
func WithLogWriter(ctx context.Context, w io.Writer) context.Context {
	return transform.WithLogWriter(ctx, w)
}
