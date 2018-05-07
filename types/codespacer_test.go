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
	// skip 1
	code4 := codespacer.RegisterNext(CodespaceType(6))
	require.Equal(t, code4, CodespaceType(6))
	code5 := codespacer.RegisterNext(CodespaceType(2))
	require.Equal(t, code5, CodespaceType(5))
	code6 := codespacer.RegisterNext(CodespaceType(2))
	require.Equal(t, code6, CodespaceType(7))
	// panic on maximum
	defer func() {
		r := recover()
		require.NotNil(t, r, "Did not panic on maximum codespace")
	}()
	codespacer.RegisterNext(MaximumCodespace - 1)
	codespacer.RegisterNext(MaximumCodespace - 1)
}

func TestRegisterOrPanic(t *testing.T) {
	codespacer := NewCodespacer()
	// unregistered, allow
	code1 := codespacer.RegisterNext(CodespaceType(2))
	require.Equal(t, code1, CodespaceType(2))
	// panic on duplicate
	defer func() {
		r := recover()
		require.NotNil(t, r, "Did not panic on duplicate codespace")
	}()
	codespacer.RegisterOrPanic(CodespaceType(2))
}
