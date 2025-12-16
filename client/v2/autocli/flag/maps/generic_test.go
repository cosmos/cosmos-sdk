package maps

import (
	"maps"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenericMapValue_Set(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		initialMap  map[string]int
		expectMap   map[string]int
		expectError bool
		changed     bool
	}{
		{
			name:       "basic key-value pairs",
			input:      "key1=1,key2=2",
			initialMap: map[string]int{},
			expectMap:  map[string]int{"key1": 1, "key2": 2},
		},
		{
			name:        "invalid format missing value",
			input:       "key1",
			initialMap:  map[string]int{},
			expectError: true,
		},
		{
			name:        "invalid format wrong separator",
			input:       "key1:1",
			initialMap:  map[string]int{},
			expectError: true,
		},
		{
			name:       "overwrite existing map first time",
			input:      "key3=3",
			initialMap: map[string]int{"key1": 1, "key2": 2},
			expectMap:  map[string]int{"key3": 3},
		},
		{
			name:        "invalid value format",
			input:       "key1=invalid",
			initialMap:  map[string]int{},
			expectError: true,
		},
		{
			name:        "empty string input",
			input:       "",
			initialMap:  map[string]int{},
			expectError: true,
		},
		{
			name:        "empty value",
			input:       "key=",
			initialMap:  map[string]int{},
			expectError: true,
		},
		{
			name:       "append to existing map",
			input:      "key3=3",
			initialMap: map[string]int{"key1": 1, "key2": 2},
			expectMap:  map[string]int{"key1": 1, "key2": 2, "key3": 3},
			changed:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a map value with string keys and int values
			mapVal := make(map[string]int)
			maps.Copy(mapVal, tc.initialMap)

			gm := newGenericMapValue(mapVal, &mapVal)
			gm.changed = tc.changed
			gm.Options = genericMapValueOptions[string, int]{
				keyParser: func(s string) (string, error) {
					return s, nil
				},
				valueParser: strconv.Atoi,
				genericType: "map[string]int",
			}

			err := gm.Set(tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectMap, mapVal)
		})
	}
}

func TestGenericMapValue_Changed(t *testing.T) {
	mapVal := make(map[string]int)
	gm := newGenericMapValue(mapVal, &mapVal)
	gm.Options = genericMapValueOptions[string, int]{
		keyParser: func(s string) (string, error) {
			return s, nil
		},
		valueParser: strconv.Atoi,
		genericType: "map[string]int",
	}

	require.False(t, gm.changed)

	// First Set should replace the map entirely
	err := gm.Set("key1=1")
	require.NoError(t, err)
	require.True(t, gm.changed)
	require.Equal(t, map[string]int{"key1": 1}, mapVal)

	// Second Set should merge with existing map
	err = gm.Set("key2=2")
	require.NoError(t, err)
	require.Equal(t, map[string]int{"key1": 1, "key2": 2}, mapVal)
}
