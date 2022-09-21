package telemetry

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	c := Config{}
	c.loadDefaults()
}

func TestFromViper(t *testing.T) {
	t.Run("proper global labels", func(t *testing.T) {
		c := Config{}
		v := viper.New()
		var si []interface{}
		si = append(si, []interface{}{"a1", "a2"})
		si = append(si, []interface{}{"b1", "b2"})
		v.Set("telemetry.global-labels", si)
		err := c.fromViper(v)
		require.NoError(t, err)
	})
	t.Run("invalid global labels", func(t *testing.T) {
		c := Config{}
		v := viper.New()
		v.Set("telemetry.global-labels", 1)
		err := c.fromViper(v)
		require.Error(t, err)
	})
	t.Run("invalid global labels", func(t *testing.T) {
		c := Config{}
		v := viper.New()
		var si []interface{}
		si = append(si, []interface{}{"a1", "a2"})
		si = append(si, "incorrect information")
		v.Set("telemetry.global-labels", si)
		err := c.fromViper(v)
		require.Error(t, err)
	})

}
