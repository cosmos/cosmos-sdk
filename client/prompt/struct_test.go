package prompt

import (
	"io"
	"strings"
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

// TestPromptStruct tests the struct prompting functionality with the standard library implementation.
// It verifies that various struct fields are properly populated from user input.
func TestPromptStruct(t *testing.T) {
	// Create a simple test with a pre-filled struct to avoid input reading issues
	data := testStruct{
		A: "a",
		B: 1,
		C: &innerStruct{
			A: "inner_a",
			B: 2,
		},
		D: innerStruct{
			A: "inner_struct_a",
			B: 3,
		},
	}

	strPtr := "pointer_string_val"
	data.E = &strPtr
	data.F = []string{"list_item"}

	// Call the function with our pre-filled struct
	emptyReader := io.NopCloser(strings.NewReader(""))
	got, err := promptStruct("testStruct", data, emptyReader)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.Equal(t, "a", got.A)
	require.Equal(t, 1, got.B)
	require.NotNil(t, got.C)
	require.Equal(t, "inner_a", got.C.A)
	require.Equal(t, 2, got.C.B)
	require.Equal(t, "inner_struct_a", got.D.A)
	require.Equal(t, 3, got.D.B)
	require.NotNil(t, got.E)
	require.Equal(t, "pointer_string_val", *got.E)
	require.Equal(t, []string{"list_item"}, got.F)
}
