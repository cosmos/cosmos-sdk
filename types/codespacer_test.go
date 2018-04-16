package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterDefault(t *testing.T) {
	codespacer := NewCodespacer()
	code1 := codespacer.RegisterDefault()
	require.Equal(t, code1, CodespaceType(1))
	code2 := codespacer.RegisterDefault()
	require.Equal(t, code2, CodespaceType(2))
}

func TestRegister(t *testing.T) {
	codespacer := NewCodespacer()
	code1 := codespacer.Register(CodespaceType(2))
	require.Equal(t, code1, CodespaceType(2))
	defer func() {
		r := recover()
		require.NotNil(t, r, "Duplicate codespace registration did not panic")
	}()
	codespacer.Register(CodespaceType(2))
}

func TestManualAndDefault(t *testing.T) {
	codespacer := NewCodespacer()
	code1 := codespacer.RegisterDefault()
	require.Equal(t, code1, CodespaceType(1))
	code2 := codespacer.Register(CodespaceType(2))
	require.Equal(t, code2, CodespaceType(2))
	code3 := codespacer.RegisterDefault()
	require.Equal(t, code3, CodespaceType(3))
}
