package viperutils_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/viperutils"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	testEnvPrefix = "TESTD"
)

func TestInitiateViper(t *testing.T) {
	testCases := map[string]struct {
		configFilesContent   []string
		environmentVariables map[string]string
		addFlags             func(*cobra.Command)
		args                 []string
		expectedValues       map[string]any
	}{
		"flag defaults are read in": {
			addFlags: func(cmd *cobra.Command) {
				cmd.Flags().String("flaggy", "from-flag", "")
			},
			expectedValues: map[string]any{
				"flaggy": "from-flag",
			},
		},
		"config file override flag default": {
			configFilesContent: []string{
				"flaggy = \"from-config-file\"",
			},
			addFlags: func(cmd *cobra.Command) {
				cmd.Flags().String("flaggy", "from-flag", "")
			},
			expectedValues: map[string]any{
				"flaggy": "from-config-file",
			},
		},
		"multiple config files merge": {
			configFilesContent: []string{
				`
flaggy = "from-config-file-1"
flaggy-2 = "also-from-config-file-1"
`, `
flaggy = "overriden-from-config-file-2"
`,
			},
			addFlags: func(cmd *cobra.Command) {
				cmd.Flags().String("flaggy", "from-flag", "")
				cmd.Flags().String("flaggy-2", "from-flag", "")
			},
			expectedValues: map[string]any{
				"flaggy":   "overriden-from-config-file-2",
				"flaggy-2": "also-from-config-file-1",
			},
		},
		"environment override flag default": {
			environmentVariables: map[string]string{
				"TESTD_FLAGGY": "from-env",
			},
			addFlags: func(cmd *cobra.Command) {
				cmd.Flags().String("flaggy", "from-flag", "")
			},
			expectedValues: map[string]any{
				"flaggy": "from-env",
			},
		},
		"env override config file": {
			environmentVariables: map[string]string{
				"TESTD_FLAGGY": "from-env",
			},
			configFilesContent: []string{
				"flaggy = \"from-config-file\"",
			},
			addFlags: func(cmd *cobra.Command) {
				cmd.Flags().String("flaggy", "from-flag", "")
			},
			expectedValues: map[string]any{
				"flaggy": "from-env",
			},
		},
		"flag override everything": {
			configFilesContent: []string{
				"flaggy = \"from config file\"",
			},
			environmentVariables: map[string]string{
				"TESTD_FLAGGY": "from env variable",
			},
			addFlags: func(cmd *cobra.Command) {
				cmd.Flags().String("flaggy", "from-flag", "")
			},
			args: []string{"--flaggy", "from-flag-override"},
			expectedValues: map[string]any{
				"flaggy": "from-flag-override",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			for key, value := range tc.environmentVariables {
				require.NoError(t, os.Setenv(key, value))
			}
			defer func() {
				for key, _ := range tc.environmentVariables {
					_ = os.Unsetenv(key)
				}
			}()

			v := viper.New()
			actual := make(map[string]any)
			cmd := &cobra.Command{
				Use: "testcommand",
				Run: func(_ *cobra.Command, _ []string) {
					for key, value := range tc.expectedValues {
						switch value.(type) {
						case string:
							actual[key] = v.GetString(key)
						case int:
							actual[key] = v.GetInt(key)
						case float64:
							actual[key] = v.GetFloat64(key)
						case bool:
							actual[key] = v.GetBool(key)
						default:
							require.Fail(t, "%v was not a predefined type in the test scenario, add the type", value)
						}
					}
				},
			}
			tc.addFlags(cmd)
			err := viperutils.InitiateViper(v, cmd, testEnvPrefix)
			require.NoError(t, err)

			configFolder := t.TempDir()
			for i, configContent := range tc.configFilesContent {
				configFile := path.Join(configFolder, fmt.Sprintf("config_file_%d.toml", i))
				err := os.WriteFile(configFile, []byte(configContent), 0o600)
				require.NoError(t, err)

				v.SetConfigFile(configFile)
				err = v.MergeInConfig()
				require.NoError(t, err)
			}

			_, err = clitestutil.ExecTestCLICmd(client.Context{}, cmd, tc.args)
			require.NoError(t, err)
			for key, value := range tc.expectedValues {
				require.Equal(t, value, actual[key])
			}

		})
	}
}

const testTemplate = `
#This is a comment
value = "{{ .Value1 }}"
another-value = "{{ .Value2 }}"
`

func TestWriteConfigToFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFilePath := path.Join(tmpDir, "config.toml")
	err := viperutils.WriteConfigToFile(configFilePath, testTemplate, struct {
		Value1 string
		Value2 string
	}{
		Value1: "test1",
		Value2: "test2",
	})
	require.NoError(t, err)

	b, err := os.ReadFile(configFilePath)
	require.NoError(t, err)
	require.Equal(t, `
#This is a comment
value = "test1"
another-value = "test2"
`, string(b))
}

type TestConfig struct {
	Key1 string
	Key2 string
	Key3 int
}

func TestGetConfig(t *testing.T) {
	v := viper.New()
	v.Set("key1", "value1")
	v.Set("key2", "value2")
	v.Set("key3", 3)

	conf, err := viperutils.GetConfig[TestConfig](v)
	require.NoError(t, err)
	require.Equal(t, TestConfig{
		Key1: "value1",
		Key2: "value2",
		Key3: 3,
	}, conf)
}
