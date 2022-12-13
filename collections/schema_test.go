package collections

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNameRegex(t *testing.T) {
	require.Regexp(t, nameRegex, "a")
	require.Regexp(t, nameRegex, "ABC")
	require.Regexp(t, nameRegex, "foo1_xyz")
	require.NotRegexp(t, nameRegex, "1foo")
	require.NotRegexp(t, nameRegex, "_bar")
	require.NotRegexp(t, nameRegex, "abc-xyz")
}

func TestAddCollection(t *testing.T) {
	sk, _ := deps()

	require.Panics(t, func() {
		schema := NewSchema(sk)
		schema.addCollection(NewMap(schema, NewPrefix(2), "123", Uint64Key, Uint64Value))
	}, "invalid name")

	require.Panics(t, func() {
		schema := NewSchema(sk)
		schema.addCollection(NewMap(schema, NewPrefix(1), "abc", Uint64Key, Uint64Value))
		schema.addCollection(NewMap(schema, NewPrefix(1), "def", Uint64Key, Uint64Value))
	}, "prefix conflict")

	require.Panics(t, func() {
		schema := NewSchema(sk)
		schema.addCollection(NewMap(schema, NewPrefix(1), "abc", Uint64Key, Uint64Value))
		schema.addCollection(NewMap(schema, NewPrefix(2), "abc", Uint64Key, Uint64Value))
	}, "name conflict")
}
