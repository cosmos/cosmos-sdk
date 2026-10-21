// Package msgguard checks that the msg servers of an app call
// msgservice.CheckMsgAllowed before anything else, so a msg that reaches its
// msg server without the MsgServiceRouter gets the circuit breaker check.
package msgguard

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

// Method is one msg server method that a module registers.
type Method struct {
	// Service is the fully-qualified Msg service name.
	Service string
	// Name is the method name.
	Name string
	// TypeURL is the type URL of the method's request msg.
	TypeURL string
	// Impl is the Go type of the msg server.
	Impl string

	desc grpc.MethodDesc
	impl any
}

func (m Method) String() string {
	return fmt.Sprintf("%s/%s (%s, %s)", m.Service, m.Name, m.TypeURL, m.Impl)
}

type service struct {
	desc *grpc.ServiceDesc
	impl any
}

type captureServer struct{ services []service }

func (c *captureServer) RegisterService(sd *grpc.ServiceDesc, impl any) {
	c.services = append(c.services, service{desc: sd, impl: impl})
}

type discardServer struct{}

func (discardServer) RegisterService(*grpc.ServiceDesc, any) {}

var (
	_ gogogrpc.Server = (*captureServer)(nil)
	_ gogogrpc.Server = discardServer{}
)

// Methods registers the services of every module in mm on a fresh
// configurator and returns every Msg service method they register, sorted by
// type URL. The app the modules belong to stays unchanged.
func Methods(t *testing.T, cdc codec.Codec, mm *module.Manager) []Method {
	t.Helper()
	msgs := &captureServer{}
	cfg := module.NewConfigurator(cdc, msgs, discardServer{})
	require.NoError(t, mm.RegisterServices(cfg))

	var out []Method
	for _, svc := range msgs.services {
		for _, md := range svc.desc.Methods {
			out = append(out, Method{
				Service: svc.desc.ServiceName,
				Name:    md.MethodName,
				TypeURL: requestTypeURL(t, svc.desc, md),
				Impl:    reflect.TypeOf(svc.impl).String(),
				desc:    md,
				impl:    svc.impl,
			})
		}
	}
	slices.SortFunc(out, func(a, b Method) int { return strings.Compare(a.TypeURL, b.TypeURL) })
	return out
}

// requestTypeURL reads the request type of a generated method handler the way
// the MsgServiceRouter does: through the decoder, with an interceptor that
// never calls the msg server.
func requestTypeURL(t *testing.T, sd *grpc.ServiceDesc, md grpc.MethodDesc) string {
	t.Helper()
	var url string
	_, _ = md.Handler(nil, context.Background(), func(in any) error {
		msg, ok := in.(sdk.Msg)
		require.True(t, ok, "%s/%s: request %T is not an sdk.Msg", sd.ServiceName, md.MethodName, in)
		url = sdk.MsgTypeURL(msg)
		return nil
	}, func(context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler) (any, error) {
		return nil, nil
	})
	require.NotEmpty(t, url, "%s/%s: no request type", sd.ServiceName, md.MethodName)
	return url
}

// errBlocked is what the recording breaker returns for every lookup. The
// guard returns a breaker error as is, so a method that runs the guard first
// returns exactly this error.
var errBlocked = errors.New("msgguard: blocked by the recording circuit breaker")

// recorder is a circuit breaker that blocks every msg and records each type
// URL it is asked about.
type recorder struct{ urls []string }

func (r *recorder) IsAllowed(_ context.Context, typeURL string) (bool, error) {
	r.urls = append(r.urls, typeURL)
	return false, errBlocked
}

// guardCtx returns a Context with cb attached and no multi-store, so a msg
// server that reads or writes state before its guard panics.
func guardCtx(cb sdk.CircuitBreaker) sdk.Context {
	return sdk.Context{}.
		WithContext(context.Background()).
		WithLogger(log.NewNopLogger()).
		WithEventManager(sdk.NewEventManager()).
		WithGasMeter(storetypes.NewInfiniteGasMeter()).
		WithCircuitBreaker(cb)
}

// Call calls the msg server method directly with msg, without the
// MsgServiceRouter, the way an EVM precompile calls a msg server. msg must
// be of the method's request type.
func (m Method) Call(ctx context.Context, msg sdk.Msg) (any, error) {
	if url := sdk.MsgTypeURL(msg); url != m.TypeURL {
		return nil, fmt.Errorf("msgguard: %s cannot take a %s", m, url)
	}
	return m.desc.Handler(m.impl, ctx, func(any) error { return nil },
		func(ctx context.Context, _ any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, msg)
		})
}

// callZero runs m on a zero request msg and returns a panic as an error,
// where Call takes a msg and lets a panic propagate.
func (m Method) callZero(ctx sdk.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	_, err = m.desc.Handler(m.impl, ctx, func(any) error { return nil }, nil)
	return err
}

// RequireGuardFirst checks every Msg service method that the modules in mm
// register, except those exempt reports true for, and returns all methods,
// exempt ones included. For each checked method, called directly with a zero
// request msg:
//
//   - on a Context with a circuit breaker that blocks everything, the method
//     returns the breaker error after asking about its own type URL exactly
//     once, before any state access or validation;
//   - on a Context that marks its own type URL as checked by the router, the
//     method skips the lookup of its own type URL;
//   - on a Context that marks another type URL as checked, the method still
//     asks about its own type URL, so a mark from an outer msg never covers
//     it.
func RequireGuardFirst(t *testing.T, cdc codec.Codec, mm *module.Manager, exempt func(Method) bool) []Method {
	t.Helper()
	methods := Methods(t, cdc, mm)
	require.NotEmpty(t, methods, "no Msg service methods registered")

	for _, m := range methods {
		if exempt(m) {
			continue
		}
		t.Run(m.TypeURL, func(t *testing.T) {
			rec := &recorder{}
			err := m.callZero(guardCtx(rec))
			require.ErrorIs(t, err, errBlocked, "%s does not call msgservice.CheckMsgAllowed first", m)
			require.Equal(t, []string{m.TypeURL}, rec.urls, "%s: guard looked up the wrong type URLs", m)

			// A msg server that calls other msg servers directly, such as
			// group CreateGroupWithPolicy, still checks those inner msgs.
			rec = &recorder{}
			_ = m.callZero(guardCtx(rec).WithCheckedMsgTypeURL(m.TypeURL))
			require.NotContains(t, rec.urls, m.TypeURL, "%s: guard repeats the router's lookup", m)

			rec = &recorder{}
			err = m.callZero(guardCtx(rec).WithCheckedMsgTypeURL(m.TypeURL + "Outer"))
			require.ErrorIs(t, err, errBlocked, "%s: a mark for another msg skips the guard", m)
			require.Equal(t, []string{m.TypeURL}, rec.urls, "%s: a mark for another msg changes the lookup", m)
		})
	}
	return methods
}

// RequireAppGuarded runs RequireGuardFirst on the modules in mm, exempting
// only the msgs x/safety always allows, then RequireRoutesSwept on router. It
// logs whether each method is guarded and returns all methods.
func RequireAppGuarded(t *testing.T, cdc codec.Codec, mm *module.Manager, registry codectypes.InterfaceRegistry, router baseapp.MessageRouter) []Method {
	t.Helper()
	exempt := func(m Method) bool { return safetytypes.IsProtected(m.TypeURL) }
	methods := RequireGuardFirst(t, cdc, mm, exempt)
	RequireRoutesSwept(t, registry, router, methods)
	for _, m := range methods {
		t.Logf("guarded=%t %s", !exempt(m), m)
	}
	return methods
}

// RequireRoutesSwept checks that methods include every msg type the router
// has a handler for, so a msg service registered outside the module manager
// cannot escape the sweep.
func RequireRoutesSwept(t *testing.T, registry codectypes.InterfaceRegistry, router baseapp.MessageRouter, methods []Method) {
	t.Helper()
	swept := map[string]bool{}
	for _, m := range methods {
		swept[m.TypeURL] = true
	}
	var routed int
	for _, url := range registry.ListImplementations(sdk.MsgInterfaceProtoName) {
		if router.HandlerByTypeURL(url) == nil {
			continue
		}
		routed++
		require.True(t, swept[url], "router handles %s but the module manager sweep never reached it", url)
	}
	require.Equal(t, routed, len(methods), "the sweep found msg server methods the router does not handle")
}
