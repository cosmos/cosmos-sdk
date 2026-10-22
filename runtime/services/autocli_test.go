package services

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestAutoCLIServiceRegistrarAccumulatesServices(t *testing.T) {
	var reg autocliServiceRegistrar
	reg.RegisterService(&grpc.ServiceDesc{ServiceName: "cosmos.foo.v1.Query"}, nil)
	reg.RegisterService(&grpc.ServiceDesc{ServiceName: "cosmos.foo.v2.Query"}, nil)

	// Both services must be retained; the second one previously overwrote the first.
	require.Equal(t, []string{"cosmos.foo.v1.Query", "cosmos.foo.v2.Query"}, reg.serviceNames)
}

func TestNewServiceCommandDescriptor(t *testing.T) {
	require.Nil(t, newServiceCommandDescriptor(nil))

	// Single service stays the primary command, no sub-commands.
	single := newServiceCommandDescriptor([]string{"cosmos.bank.v1beta1.Query"})
	require.Equal(t, "cosmos.bank.v1beta1.Query", single.Service)
	require.Empty(t, single.SubCommands)

	// Multiple services: first is primary, the rest become sub-commands and are
	// no longer dropped.
	multi := newServiceCommandDescriptor([]string{
		"cosmos.foo.v1.Query",
		"cosmos.foo.v1.SecondaryQuery",
	})
	require.Equal(t, "cosmos.foo.v1.Query", multi.Service)
	require.Len(t, multi.SubCommands, 1)
	require.Contains(t, multi.SubCommands, "secondaryquery")
	require.Equal(t, "cosmos.foo.v1.SecondaryQuery", multi.SubCommands["secondaryquery"].Service)
}

func TestSubCommandKeyCollisionFallsBackToFullName(t *testing.T) {
	// Two services share the short name "Query" but differ by version, so the
	// second must fall back to its fully qualified name to stay unique.
	desc := newServiceCommandDescriptor([]string{
		"cosmos.foo.v1.Service",
		"cosmos.foo.v1.Query",
		"cosmos.foo.v2.Query",
	})
	require.Equal(t, "cosmos.foo.v1.Service", desc.Service)
	require.Len(t, desc.SubCommands, 2)
	require.Contains(t, desc.SubCommands, "query")
	require.Contains(t, desc.SubCommands, "cosmos.foo.v2.query")
	require.Equal(t, "cosmos.foo.v1.Query", desc.SubCommands["query"].Service)
	require.Equal(t, "cosmos.foo.v2.Query", desc.SubCommands["cosmos.foo.v2.query"].Service)
}
