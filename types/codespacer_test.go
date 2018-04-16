package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterNext(t *testing.T) {
	codespacer := NewCodespacer()
	// unregistered, allow
	code1 := codespacer.RegisterNext(CodespaceType(2))
	require.Equal(t, code1, CodespaceType(2))
	// registered, pick next
	code2 := codespacer.RegisterNext(CodespaceType(2))
	require.Equal(t, code2, CodespaceType(3))
	// pick next
	code3 := codespacer.RegisterNext(CodespaceType(2))
	require.Equal(t, code3, CodespaceType(4))
}
