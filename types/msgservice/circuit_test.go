package msgservice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// stubBreaker blocks the type URLs in blocked, returns err when set, and
// counts lookups. With allowedOnErr set, it reports the msg as allowed
// alongside err, as contrib/x/circuit does when its store read fails.
type stubBreaker struct {
	blocked      map[string]bool
	err          error
	allowedOnErr bool
	calls        int
}

func (b *stubBreaker) IsAllowed(_ context.Context, typeURL string) (bool, error) {
	b.calls++
	if b.err != nil {
		return b.allowedOnErr, b.err
	}
	return !b.blocked[typeURL], nil
}

func TestCheckMsgAllowed(t *testing.T) {
	msg := &testdata.TestMsg{}
	msgURL := sdk.MsgTypeURL(msg)
	errStore := errors.New("store failure")

	testCases := []struct {
		name      string
		breaker   *stubBreaker
		checked   string
		wrap      bool
		bypass    bool
		expErr    error
		expLookup int
	}{
		{name: "no breaker", expLookup: 0},
		{name: "allowed", breaker: &stubBreaker{}, expLookup: 1},
		{name: "blocked", breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}}, expErr: sdkerrors.ErrUnauthorized, expLookup: 1},
		{name: "breaker error fails closed", breaker: &stubBreaker{err: errStore}, expErr: errStore, expLookup: 1},
		{name: "breaker error fails closed when it also allows", breaker: &stubBreaker{err: errStore, allowedOnErr: true}, expErr: errStore, expLookup: 1},
		{name: "router checked this msg type", breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}}, checked: msgURL, expLookup: 0},
		{name: "router checked another msg type", breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}}, checked: "/testpb.MsgCreateDog", expErr: sdkerrors.ErrUnauthorized, expLookup: 1},
		{name: "sdk context behind a value key", breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}}, wrap: true, expErr: sdkerrors.ErrUnauthorized, expLookup: 1},
		{name: "bypass", breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}}, bypass: true, expLookup: 0},
		{name: "bypass behind a value key", breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}}, wrap: true, bypass: true, expLookup: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sdkCtx := sdk.Context{}.WithCheckedMsgTypeURL(tc.checked).WithCircuitBreakerBypass(tc.bypass)
			if tc.breaker != nil {
				sdkCtx = sdkCtx.WithCircuitBreaker(tc.breaker)
			}
			var ctx context.Context = sdkCtx
			if tc.wrap {
				ctx = context.WithValue(context.Background(), sdk.SdkContextKey, sdkCtx)
			}

			err := msgservice.CheckMsgAllowed(ctx, msg)
			if tc.expErr != nil {
				require.ErrorIs(t, err, tc.expErr)
			} else {
				require.NoError(t, err)
			}
			if tc.breaker != nil {
				require.Equal(t, tc.expLookup, tc.breaker.calls)
			}
		})
	}
}

func TestCheckMsgAllowedWithoutSDKContext(t *testing.T) {
	require.NoError(t, msgservice.CheckMsgAllowed(context.Background(), &testdata.TestMsg{}))
}

// TestCheckCircuitBreaker covers the check the MsgServiceRouter runs, which
// must fail closed on a breaker error whatever the breaker reports.
func TestCheckCircuitBreaker(t *testing.T) {
	msgURL := sdk.MsgTypeURL(&testdata.TestMsg{})
	errStore := errors.New("store failure")

	testCases := []struct {
		name      string
		breaker   *stubBreaker
		ctx       context.Context
		expErr    error
		expLookup int
	}{
		{name: "allowed", breaker: &stubBreaker{}, expLookup: 1},
		{name: "blocked", breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}}, expErr: sdkerrors.ErrUnauthorized, expLookup: 1},
		{name: "breaker error", breaker: &stubBreaker{err: errStore}, expErr: errStore, expLookup: 1},
		{name: "breaker error when it also allows", breaker: &stubBreaker{err: errStore, allowedOnErr: true}, expErr: errStore, expLookup: 1},
		{name: "bypass", breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}}, ctx: sdk.Context{}.WithCircuitBreakerBypass(true)},
		{
			name:    "bypass behind a value key",
			breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}},
			ctx:     context.WithValue(context.Background(), sdk.SdkContextKey, sdk.Context{}.WithCircuitBreakerBypass(true)),
		},
		{name: "no sdk context", breaker: &stubBreaker{blocked: map[string]bool{msgURL: true}}, ctx: context.Background(), expErr: sdkerrors.ErrUnauthorized, expLookup: 1},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.ctx
			if ctx == nil {
				ctx = sdk.Context{}
			}
			err := msgservice.CheckCircuitBreaker(ctx, tc.breaker, msgURL)
			if tc.expErr != nil {
				require.ErrorIs(t, err, tc.expErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expLookup, tc.breaker.calls)
		})
	}
}
