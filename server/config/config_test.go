package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.True(t, cfg.GetMinGasPrices().IsZero())
}

func TestSetMinimumFees(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SetMinGasPrices(sdk.DecCoins{sdk.NewInt64DecCoin("foo", 5)})
	require.Equal(t, "5.000000000000000000foo", cfg.MinGasPrices)
}

func TestIndexEventsMarshalling(t *testing.T) {
	expectedIn := `index-events = ["key1", "key2", ]` + "\n"
	cfg := DefaultConfig()
	cfg.IndexEvents = []string{"key1", "key2"}
	var buffer bytes.Buffer

	err := configTemplate.Execute(&buffer, cfg)
	require.NoError(t, err, "executing template")
	actual := buffer.String()
	require.Contains(t, actual, expectedIn, "config file contents")
}

func TestStreamingConfig(t *testing.T) {
	cfg := Config{
		Streaming: StreamingConfig{
			ABCI: ABCIListenerConfig{
				Keys:          []string{"one", "two"},
				Plugin:        "plugin-A",
				StopNodeOnErr: false,
			},
		},
	}

	testDir := t.TempDir()
	cfgFile := filepath.Join(testDir, "app.toml")
	WriteConfigFile(cfgFile, &cfg)

	cfgFileBz, err := os.ReadFile(cfgFile)
	require.NoError(t, err, "reading %s", cfgFile)
	cfgFileContents := string(cfgFileBz)
	t.Logf("Config file contents: %s:\n%s", cfgFile, cfgFileContents)

	expectedLines := []string{
		`keys = ["one", "two", ]`,
		`plugin = "plugin-A"`,
		`stop-node-on-err = false`,
	}

	for _, line := range expectedLines {
		assert.Contains(t, cfgFileContents, line+"\n", "config file contents")
	}

	vpr := viper.New()
	vpr.SetConfigFile(cfgFile)
	err = vpr.ReadInConfig()
	require.NoError(t, err, "reading config file into viper")

	var actual Config
	err = vpr.Unmarshal(&actual)
	require.NoError(t, err, "vpr.Unmarshal")

	assert.Equal(t, cfg.Streaming, actual.Streaming, "Streaming")
}

func TestParseStreaming(t *testing.T) {
	expectedKeys := `keys = ["*", ]` + "\n"
	expectedPlugin := `plugin = "abci_v1"` + "\n"
	expectedStopNodeOnErr := `stop-node-on-err = true` + "\n"

	cfg := DefaultConfig()
	cfg.Streaming.ABCI.Keys = []string{"*"}
	cfg.Streaming.ABCI.Plugin = "abci_v1"
	cfg.Streaming.ABCI.StopNodeOnErr = true

	var buffer bytes.Buffer
	err := configTemplate.Execute(&buffer, cfg)
	require.NoError(t, err, "executing template")
	actual := buffer.String()
	require.Contains(t, actual, expectedKeys, "config file contents")
	require.Contains(t, actual, expectedPlugin, "config file contents")
	require.Contains(t, actual, expectedStopNodeOnErr, "config file contents")
}

func TestReadConfig(t *testing.T) {
	cfg := DefaultConfig()
	tmpFile := filepath.Join(t.TempDir(), "config")
	WriteConfigFile(tmpFile, cfg)

	v := viper.New()
	otherCfg, err := GetConfig(v)
	require.NoError(t, err)

	require.Equal(t, *cfg, otherCfg)
}

func TestIndexEventsWriteRead(t *testing.T) {
	expected := []string{"key3", "key4"}

	// Create config with two IndexEvents entries, and write it to a file.
	confFile := filepath.Join(t.TempDir(), "app.toml")
	conf := DefaultConfig()
	conf.IndexEvents = expected

	WriteConfigFile(confFile, conf)

	// read the file into Viper
	vpr := viper.New()
	vpr.SetConfigFile(confFile)

	err := vpr.ReadInConfig()
	require.NoError(t, err, "reading config file into viper")

	// Check that the raw viper value is correct.
	actualRaw := vpr.GetStringSlice("index-events")
	require.Equal(t, expected, actualRaw, "viper's index events")

	// Check that it is parsed into the config correctly.
	cfg, perr := ParseConfig(vpr)
	require.NoError(t, perr, "parsing config")

	actual := cfg.IndexEvents
	require.Equal(t, expected, actual, "config value")
}

func TestGlobalLabelsEventsMarshalling(t *testing.T) {
	expectedIn := `global-labels = [
  ["labelname1", "labelvalue1"],
  ["labelname2", "labelvalue2"],
]`
	cfg := DefaultConfig()
	cfg.Telemetry.GlobalLabels = [][]string{{"labelname1", "labelvalue1"}, {"labelname2", "labelvalue2"}}
	var buffer bytes.Buffer

	err := configTemplate.Execute(&buffer, cfg)
	require.NoError(t, err, "executing template")
	actual := buffer.String()
	require.Contains(t, actual, expectedIn, "config file contents")
}

func TestGlobalLabelsWriteRead(t *testing.T) {
	expected := [][]string{{"labelname3", "labelvalue3"}, {"labelname4", "labelvalue4"}}
	expectedRaw := make([]interface{}, len(expected))
	for i, exp := range expected {
		pair := make([]interface{}, len(exp))
		for j, s := range exp {
			pair[j] = s
		}
		expectedRaw[i] = pair
	}

	// Create config with two GlobalLabels entries, and write it to a file.
	confFile := filepath.Join(t.TempDir(), "app.toml")
	conf := DefaultConfig()
	conf.Telemetry.GlobalLabels = expected
	WriteConfigFile(confFile, conf)

	// Read that file into viper.
	vpr := viper.New()
	vpr.SetConfigFile(confFile)
	rerr := vpr.ReadInConfig()
	require.NoError(t, rerr, "reading config file into viper")
	// Check that the raw viper value is correct.
	actualRaw := vpr.Get("telemetry.global-labels")
	require.Equal(t, expectedRaw, actualRaw, "viper value")
	// Check that it is parsed into the config correctly.
	cfg, perr := ParseConfig(vpr)
	require.NoError(t, perr, "parsing config")
	actual := cfg.Telemetry.GlobalLabels
	require.Equal(t, expected, actual, "config value")
}

func TestSetConfigTemplate(t *testing.T) {
	conf := DefaultConfig()
	var initBuffer, setBuffer bytes.Buffer

	// Use the configTemplate defined during init() to create a config string.
	ierr := configTemplate.Execute(&initBuffer, conf)
	require.NoError(t, ierr, "initial configTemplate.Execute")
	expected := initBuffer.String()

	// Set the template to the default one.
	initTmpl := configTemplate
	require.NotPanics(t, func() {
		SetConfigTemplate(DefaultConfigTemplate)
	}, "SetConfigTemplate")
	setTmpl := configTemplate
	require.NotSame(t, initTmpl, setTmpl, "configTemplate after set")

	// Create the string again and make sure it's the same.
	serr := configTemplate.Execute(&setBuffer, conf)
	require.NoError(t, serr, "after SetConfigTemplate, configTemplate.Execute")
	actual := setBuffer.String()
	require.Equal(t, expected, actual, "resulting config strings")
}
