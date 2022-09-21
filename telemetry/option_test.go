package telemetry

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestOptionWithConfig(t *testing.T) {
	c := &Config{}
	err := OptionWithConfig(Config{})(c)
	require.NoError(t, err)
}

func TestOptionFromViper(t *testing.T) {
	c := &Config{}
	v := viper.New()
	v.SetDefault("telemetry.global-labels", []interface{}{})
	err := OptionFromViper(v)(c)
	require.NoError(t, err)
}

func TestOptionWithEnable(t *testing.T) {
	c := &Config{}
	err := OptionWithEnable(true)(c)
	require.NoError(t, err)
}

func TestOptionWithServiceName(t *testing.T) {
	c := &Config{}
	err := OptionWithServiceName("some_service_name")(c)
	require.NoError(t, err)
}

func TestOptionWithEnableHostname(t *testing.T) {
	c := &Config{}
	err := OptionWithEnableHostname(true)(c)
	require.NoError(t, err)
}

func TestOptionWithEnableHostnameLabel(t *testing.T) {
	c := &Config{}
	err := OptionWithEnableHostnameLabel(true)(c)
	require.NoError(t, err)
}

func TestOptionWithEnableServiceLabel(t *testing.T) {
	c := &Config{}
	err := OptionWithEnableServiceLabel(true)(c)
	require.NoError(t, err)
}

func TestOptionWithPrometheusRetentionTime(t *testing.T) {
	c := &Config{}
	err := OptionWithPrometheusRetentionTime(1 * time.Second)(c)
	require.NoError(t, err)
}

func TestOptionWithGlobalLabels(t *testing.T) {
	c := &Config{}
	err := OptionWithGlobalLabels([][]string{})(c)
	require.NoError(t, err)
}
