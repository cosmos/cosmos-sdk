package testutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
)

// ProtoDeepEqual is a helper function that uses the protocmp package to compare two protobuf messages.
func ProtoDeepEqual(t *testing.T, p1, p2 interface{}) {
	t.Helper()
	require.Empty(t, cmp.Diff(p1, p2, protocmp.Transform()))
}
