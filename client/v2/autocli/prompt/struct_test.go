package prompt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type innerStruct struct {
	A string
	B int
}

type testStruct struct {
	A string
	B int
	C *innerStruct
	D innerStruct
	E *string
	F []string
}

func TestPromptStruct(t *testing.T) {
	type testCase[T any] struct {
		name   string
		data   T
		inputs []string
	}
	tests := []testCase[testStruct]{
		{
			name: "test struct",
			data: testStruct{},
			inputs: []string{
				"a", "1", "b", "2", "c", "3", "pointerStr", "list",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs := getReader(tt.inputs)
			got, err := promptStruct("testStruct", tt.data, inputs)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
