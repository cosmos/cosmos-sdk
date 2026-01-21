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

func TestGetAndSetMinimumGas(t *testing.T) {
	cfg := DefaultConfig()

	input := sdk.DecCoins{sdk.NewInt64DecCoin("foo", 5)}
	cfg.SetMinGasPrices(input)
	require.Equal(t, "5.000000000000000000foo", cfg.MinGasPrices)
	require.EqualValues(t, cfg.GetMinGasPrices(), input)

	input = sdk.DecCoins{sdk.NewInt64DecCoin("bar", 1), sdk.NewInt64DecCoin("foo", 5)}
	cfg.SetMinGasPrices(input)
	require.Equal(t, "1.000000000000000000bar,5.000000000000000000foo", cfg.MinGasPrices)
	require.EqualValues(t, cfg.GetMinGasPrices(), input)
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
	expectedRaw := make([]any, len(expected))
	for i, exp := range expected {
		pair := make([]any, len(exp))
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

func TestAppConfig(t *testing.T) {
	appConfigFile := filepath.Join(t.TempDir(), "app.toml")
	defer func() {
		_ = os.Remove(appConfigFile)
	}()

	defAppConfig := DefaultConfig()
	SetConfigTemplate(DefaultConfigTemplate)
	WriteConfigFile(appConfigFile, defAppConfig)

	v := viper.New()
	v.SetConfigFile(appConfigFile)
	require.NoError(t, v.ReadInConfig())
	appCfg := new(Config)
	require.NoError(t, v.Unmarshal(appCfg))
	require.EqualValues(t, appCfg, defAppConfig)
}

func TestGetConfig_HistoricalGRPCAddressBlockRange(t *testing.T) {
	tests := []struct {
		name        string
		setupViper  func(*viper.Viper)
		expectError bool
		errorMsg    string
		validate    func(*testing.T, Config)
	}{
		{
			name: "valid single historical grpc address",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [0, 1000]}`)
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				t.Helper()
				require.Len(t, cfg.GRPC.HistoricalGRPCAddressBlockRange, 1)
				expectedRange := BlockRange{0, 1000}
				address, exists := cfg.GRPC.HistoricalGRPCAddressBlockRange[expectedRange]
				require.True(t, exists, "Block range [0, 1000] should exist")
				require.Equal(t, "localhost:9091", address)
			},
		},
		{
			name: "valid multiple historical grpc addresses",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range",
					`{"localhost:9091": [0, 1000], "localhost:9092": [1001, 2000], "localhost:9093": [2001, 3000]}`)
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				t.Helper()
				require.Len(t, cfg.GRPC.HistoricalGRPCAddressBlockRange, 3)
				testCases := []struct {
					blockRange BlockRange
					address    string
				}{
					{BlockRange{0, 1000}, "localhost:9091"},
					{BlockRange{1001, 2000}, "localhost:9092"},
					{BlockRange{2001, 3000}, "localhost:9093"},
				}
				for _, tc := range testCases {
					address, exists := cfg.GRPC.HistoricalGRPCAddressBlockRange[tc.blockRange]
					require.True(t, exists, "Block range %v should exist", tc.blockRange)
					require.Equal(t, tc.address, address)
				}
			},
		},
		{
			name: "empty configuration",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", "")
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				t.Helper()
				require.Nil(t, cfg.GRPC.HistoricalGRPCAddressBlockRange)
			},
		},
		{
			name:        "no configuration set",
			setupViper:  func(v *viper.Viper) {},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				t.Helper()
				require.Nil(t, cfg.GRPC.HistoricalGRPCAddressBlockRange)
			},
		},
		{
			name: "invalid JSON format",
			setupViper: func(v *viper.Viper) {
				// missing closing brace
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [0, 1000]`)
			},
			expectError: true,
			errorMsg:    "failed to parse historical-grpc-address-block-range as JSON",
		},
		{
			name: "negative start block",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [-1, 1000]}`)
			},
			expectError: true,
			errorMsg:    "block numbers cannot be negative",
		},
		{
			name: "negative end block",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [0, -100]}`)
			},
			expectError: true,
			errorMsg:    "block numbers cannot be negative",
		},
		{
			name: "start block greater than end block",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [1000, 500]}`)
			},
			expectError: true,
			errorMsg:    "start block must be <= end block",
		},
		{
			name: "single block range (start equals end)",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [1000, 1000]}`)
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				t.Helper()
				require.Len(t, cfg.GRPC.HistoricalGRPCAddressBlockRange, 1)
				expectedRange := BlockRange{1000, 1000}
				address, exists := cfg.GRPC.HistoricalGRPCAddressBlockRange[expectedRange]
				require.True(t, exists)
				require.Equal(t, "localhost:9091", address)
			},
		},
		{
			name: "zero to zero range",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [0, 0]}`)
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				t.Helper()
				require.Len(t, cfg.GRPC.HistoricalGRPCAddressBlockRange, 1)
				expectedRange := BlockRange{0, 0}
				address, exists := cfg.GRPC.HistoricalGRPCAddressBlockRange[expectedRange]
				require.True(t, exists)
				require.Equal(t, "localhost:9091", address)
			},
		},
		{
			name: "large block numbers",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [1000000, 2000000]}`)
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				t.Helper()
				require.Len(t, cfg.GRPC.HistoricalGRPCAddressBlockRange, 1)
				expectedRange := BlockRange{1000000, 2000000}
				address, exists := cfg.GRPC.HistoricalGRPCAddressBlockRange[expectedRange]
				require.True(t, exists)
				require.Equal(t, "localhost:9091", address)
			},
		},
		{
			name: "invalid array length (too few elements)",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [100]}`)
			},
			expectError: true,
			errorMsg:    "start block must be <= end block",
		},
		{
			name: "invalid array length (too many elements)",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"localhost:9091": [0, 1000, 2000]}`)
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				t.Helper()
				require.Len(t, cfg.GRPC.HistoricalGRPCAddressBlockRange, 1)
				expectedRange := BlockRange{0, 1000}
				address, exists := cfg.GRPC.HistoricalGRPCAddressBlockRange[expectedRange]
				require.True(t, exists)
				require.Equal(t, "localhost:9091", address)
			},
		},
		{
			name: "address with port and protocol",
			setupViper: func(v *viper.Viper) {
				v.Set("grpc.historical-grpc-address-block-range", `{"https://archive.example.com:9091": [0, 1000]}`)
			},
			expectError: false,
			validate: func(t *testing.T, cfg Config) {
				t.Helper()
				require.Len(t, cfg.GRPC.HistoricalGRPCAddressBlockRange, 1)
				expectedRange := BlockRange{0, 1000}
				address, exists := cfg.GRPC.HistoricalGRPCAddressBlockRange[expectedRange]
				require.True(t, exists)
				require.Equal(t, "https://archive.example.com:9091", address)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			v.Set("minimum-gas-prices", "0stake")
			tt.setupViper(v)
			cfg, err := GetConfig(v)
			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

func TestConfigTemplate_HistoricalGRPCAddressBlockRange(t *testing.T) {
	tests := []struct {
		name     string
		config   map[BlockRange]string
		expected string
	}{
		{
			name:     "empty config",
			config:   nil,
			expected: `historical-grpc-address-block-range = "{}"`,
		},
		{
			name: "single entry",
			config: map[BlockRange]string{
				{0, 1000}: "localhost:9091",
			},
			expected: `historical-grpc-address-block-range = "{\"localhost:9091\": [0, 1000]}"`,
		},
		{
			name: "multiple entries",
			config: map[BlockRange]string{
				{0, 1000}:    "localhost:9091",
				{1001, 2000}: "localhost:9092",
				{2001, 3000}: "localhost:9093",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.GRPC.HistoricalGRPCAddressBlockRange = tt.config

			var buffer bytes.Buffer
			err := configTemplate.Execute(&buffer, cfg)
			require.NoError(t, err)

			output := buffer.String()
			if tt.expected != "" {
				require.Contains(t, output, tt.expected)
			}

			v := viper.New()
			v.SetConfigType("toml")
			require.NoError(t, v.ReadConfig(&buffer))

			parsedCfg, err := GetConfig(v)
			require.NoError(t, err)

			if tt.config == nil {
				require.Empty(t, parsedCfg.GRPC.HistoricalGRPCAddressBlockRange)
			} else {
				require.Equal(t, len(tt.config), len(parsedCfg.GRPC.HistoricalGRPCAddressBlockRange))
				for blockRange, address := range tt.config {
					parsedAddr, exists := parsedCfg.GRPC.HistoricalGRPCAddressBlockRange[blockRange]
					require.True(t, exists, "Block range %v should exist", blockRange)
					require.Equal(t, address, parsedAddr)
				}
			}
		})
	}
}
