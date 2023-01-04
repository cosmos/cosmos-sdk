package confix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/creachadair/atomicfile"
	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/transform"
	"github.com/spf13/viper"

	srvcfg "github.com/cosmos/cosmos-sdk/server/config"
)

// Upgrade reads the configuration file at configPath and applies any
// transformations necessary to upgrade it to the current version.  If this
// succeeds, the transformed output is written to outputPath. As a special
// case, if outputPath == "" the output is written to stdout.
//
// It is safe if outputPath == inputPath. If a regular file outputPath already
// exists, it is overwritten. In case of error, the output is not written.
//
// Upgrade is a convenience wrapper for calls to LoadConfig, ApplyFixes, and
// CheckValid. If the caller requires more control over the behavior of the
// upgrade, call those functions directly.
func Upgrade(ctx context.Context, configPath, outputPath string, force bool) error {
	if configPath == "" {
		return errors.New("empty input configuration path")
	}

	doc, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %v", err)
	}

	if err := ApplyFixes(ctx, doc, force); err != nil {
		return fmt.Errorf("updating %q: %v", configPath, err)
	}

	var buf bytes.Buffer
	if err := tomledit.Format(&buf, doc); err != nil {
		return fmt.Errorf("formatting config: %v", err)
	}

	// verify that file is valid after applying fixes
	if err := CheckValid(buf.Bytes()); err != nil {
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
func ApplyFixes(ctx context.Context, doc *tomledit.Document, force bool) error {
	if sdkVersion := ExtractVersion(doc, force); sdkVersion == vUnknown && !force {
		return errors.New("unknown SDK version")
	}

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

const (
	vUnknown = ""
	v45      = "0.45"
	v46      = "0.46"
	v47      = "0.47"
	vNext    = "next" // unreleased version of the SDK
)

// ExtractVersion extracts the version of Cosmos SDK that created the config.
// If the version cannot be determined, it returns vUnknown.
// When force is true, it returns vNext. This allows the user to use the tool on latest unreleased SDK version.
func ExtractVersion(doc *tomledit.Document, force bool) string {
	if force {
		return vNext
	}

	version := doc.Find("version")
	if len(version) == 0 {
		return vUnknown
	}

	versionValue := version[0].Value.String()
	// remove quotes
	versionValue = strings.Trim(versionValue, "\"")
	switch {
	case strings.HasPrefix(versionValue, v45):
		return v45
	case strings.HasPrefix(versionValue, v46):
		return v46
	case strings.HasPrefix(versionValue, v47):
		return v47
	default:
		return vUnknown
	}
}

// CheckValid checks whether the specified config appears to be a valid
// Cosmos SDK config file. This emulates how the node loads the config.
func CheckValid(data []byte) error {
	v := viper.New()
	v.SetConfigType("toml")

	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	var cfg srvcfg.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("decoding config: %w", err)
	}

	return cfg.ValidateBasic()
}

// WithLogWriter returns a child of ctx with a logger attached that sends
// output to w. This is a convenience wrapper for transform.WithLogWriter.
func WithLogWriter(ctx context.Context, w io.Writer) context.Context {
	return transform.WithLogWriter(ctx, w)
}
