package coretesting

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/server"
)

func Test_ConfigMap_SubConfig(t *testing.T) {
	config := server.ConfigMap{
		"parent": map[string]any{
			"child": map[string]any{
				"key": "value",
			},
		},
	}

	subConfig := config.Get("parent.child")
	require.NotNil(t, subConfig)
	m := subConfig.(server.ConfigMap)
	require.Equal(t, "value", m["key"])

	nonExistentSubConfig := config.Get("parent.nonexistent.key")
	require.Nil(t, nonExistentSubConfig)
}
