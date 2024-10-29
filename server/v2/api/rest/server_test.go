package rest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/transaction"
)

func TestServerConfig(t *testing.T) {
	testCases := []struct {
		name           string
		setupFunc      func() *Config
		expectedConfig *Config
	}{
		{
			name: "Default configuration, no custom configuration",
			setupFunc: func() *Config {
				s := &Server[transaction.Tx]{}
				return s.Config().(*Config)
			},
			expectedConfig: DefaultConfig(),
		},
		{
			name: "Custom configuration",
			setupFunc: func() *Config {
				s := NewWithConfigOptions[transaction.Tx](func(config *Config) {
					config.Enable = false
				})
				return s.Config().(*Config)
			},
			expectedConfig: &Config{
				Enable:  false, // Custom configuration
				Address: "localhost:8080",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.setupFunc()
			require.Equal(t, tc.expectedConfig, config)
		})
	}
}
