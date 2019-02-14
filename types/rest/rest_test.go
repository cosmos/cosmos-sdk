// Package rest provides HTTP types and primitives for REST
// requests validation and responses handling.
package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types"
)

type mockResponseWriter struct{}

func TestBaseReq_ValidateBasic(t *testing.T) {
	tenstakes, err := types.ParseCoins("10stake")
	require.NoError(t, err)
	onestake, err := types.ParseDecCoins("1.0stake")
	require.NoError(t, err)

	req1 := NewBaseReq(
		"nonempty", "nonempty", "", "nonempty", "", "",
		0, 0, tenstakes, nil, false, false,
	)
	req2 := NewBaseReq(
		"", "nonempty", "", "nonempty", "", "",
		0, 0, tenstakes, nil, false, false,
	)
	req3 := NewBaseReq(
		"nonempty", "", "", "nonempty", "", "",
		0, 0, tenstakes, nil, false, false,
	)
	req4 := NewBaseReq(
		"nonempty", "nonempty", "", "", "", "",
		0, 0, tenstakes, nil, false, false,
	)
	req5 := NewBaseReq(
		"nonempty", "nonempty", "", "nonempty", "", "",
		0, 0, tenstakes, onestake, false, false,
	)
	req6 := NewBaseReq(
		"nonempty", "nonempty", "", "nonempty", "", "",
		0, 0, types.Coins{}, types.DecCoins{}, false, false,
	)

	tests := []struct {
		name string
		req  BaseReq
		w    http.ResponseWriter
		want bool
	}{
		{"ok", req1, httptest.NewRecorder(), true},
		{"neither fees nor gasprices provided", req6, httptest.NewRecorder(), true},
		{"empty from", req2, httptest.NewRecorder(), false},
		{"empty password", req3, httptest.NewRecorder(), false},
		{"empty chain-id", req4, httptest.NewRecorder(), false},
		{"fees and gasprices provided", req5, httptest.NewRecorder(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.req.ValidateBasic(tt.w))
		})
	}
}
