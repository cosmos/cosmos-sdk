// Package rest provides HTTP types and primitives for REST
// requests validation and responses handling.
package rest

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
)

type mockResponseWriter struct{}

func TestBaseReqValidateBasic(t *testing.T) {
	fromAddr := "cosmos1cq0sxam6x4l0sv9yz3a2vlqhdhvt2k6jtgcse0"
	tenstakes, err := types.ParseCoins("10stake")
	require.NoError(t, err)
	onestake, err := types.ParseDecCoins("1.0stake")
	require.NoError(t, err)

	req1 := NewBaseReq(
		fromAddr, "", "nonempty", "", "", 0, 0, tenstakes, nil, false,
	)
	req2 := NewBaseReq(
		"", "", "nonempty", "", "", 0, 0, tenstakes, nil, false,
	)
	req3 := NewBaseReq(
		fromAddr, "", "", "", "", 0, 0, tenstakes, nil, false,
	)
	req4 := NewBaseReq(
		fromAddr, "", "nonempty", "", "", 0, 0, tenstakes, onestake, false,
	)
	req5 := NewBaseReq(
		fromAddr, "", "nonempty", "", "", 0, 0, types.Coins{}, types.DecCoins{}, false,
	)

	tests := []struct {
		name string
		req  BaseReq
		w    http.ResponseWriter
		want bool
	}{
		{"ok", req1, httptest.NewRecorder(), true},
		{"neither fees nor gasprices provided", req5, httptest.NewRecorder(), true},
		{"empty from", req2, httptest.NewRecorder(), false},
		{"empty chain-id", req3, httptest.NewRecorder(), false},
		{"fees and gasprices provided", req4, httptest.NewRecorder(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.req.ValidateBasic(tt.w))
		})
	}
}

func TestParseHTTPArgs(t *testing.T) {
	req0 := mustNewRequest(t, "", "/", nil)
	req1 := mustNewRequest(t, "", "/?limit=5", nil)
	req2 := mustNewRequest(t, "", "/?page=5", nil)
	req3 := mustNewRequest(t, "", "/?page=5&limit=5", nil)

	reqE1 := mustNewRequest(t, "", "/?page=-1", nil)
	reqE2 := mustNewRequest(t, "", "/?limit=-1", nil)
	req4 := mustNewRequest(t, "", "/?foo=faa", nil)

	tests := []struct {
		name  string
		req   *http.Request
		w     http.ResponseWriter
		tags  []string
		page  int
		limit int
		err   bool
	}{
		{"no params", req0, httptest.NewRecorder(), []string{}, DefaultPage, DefaultLimit, false},
		{"Limit", req1, httptest.NewRecorder(), []string{}, DefaultPage, 5, false},
		{"Page", req2, httptest.NewRecorder(), []string{}, 5, DefaultLimit, false},
		{"Page and limit", req3, httptest.NewRecorder(), []string{}, 5, 5, false},

		{"error page 0", reqE1, httptest.NewRecorder(), []string{}, DefaultPage, DefaultLimit, true},
		{"error limit 0", reqE2, httptest.NewRecorder(), []string{}, DefaultPage, DefaultLimit, true},

		{"tags", req4, httptest.NewRecorder(), []string{"foo='faa'"}, DefaultPage, DefaultLimit, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, page, limit, err := ParseHTTPArgs(tt.req)
			if tt.err {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
				require.Equal(t, tt.tags, tags)
				require.Equal(t, tt.page, page)
				require.Equal(t, tt.limit, limit)
			}
		})
	}
}

func TestParseQueryHeight(t *testing.T) {
	var emptyHeight int64
	height := int64(1256756)

	req0 := mustNewRequest(t, "", "/", nil)
	req1 := mustNewRequest(t, "", "/?height=1256756", nil)
	req2 := mustNewRequest(t, "", "/?height=456yui4567", nil)
	req3 := mustNewRequest(t, "", "/?height=-1", nil)

	tests := []struct {
		name           string
		req            *http.Request
		w              http.ResponseWriter
		cliCtx         context.CLIContext
		expectedHeight int64
		expectedOk     bool
	}{
		{"no height", req0, httptest.NewRecorder(), context.CLIContext{}, emptyHeight, true},
		{"height", req1, httptest.NewRecorder(), context.CLIContext{}, height, true},
		{"invalid height", req2, httptest.NewRecorder(), context.CLIContext{}, emptyHeight, false},
		{"negative height", req3, httptest.NewRecorder(), context.CLIContext{}, emptyHeight, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliCtx, ok := ParseQueryHeightOrReturnBadRequest(tt.w, tt.cliCtx, tt.req)
			if tt.expectedOk {
				require.True(t, ok)
				require.Equal(t, tt.expectedHeight, cliCtx.Height)
			} else {
				require.False(t, ok)
				require.Empty(t, tt.expectedHeight, cliCtx.Height)
			}
		})
	}
}

func TestProcessPostResponse(t *testing.T) {
	// mock account
	// PubKey field ensures amino encoding is used first since standard
	// JSON encoding will panic on crypto.PubKey

	type mockAccount struct {
		Address       types.AccAddress `json:"address"`
		Coins         types.Coins      `json:"coins"`
		PubKey        crypto.PubKey    `json:"public_key"`
		AccountNumber uint64           `json:"account_number"`
		Sequence      uint64           `json:"sequence"`
	}

	// setup
	ctx := context.NewCLIContext()
	height := int64(194423)

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := types.AccAddress(pubKey.Address())
	coins := types.NewCoins(types.NewCoin("atom", types.NewInt(100)), types.NewCoin("tree", types.NewInt(125)))
	accNumber := uint64(104)
	sequence := uint64(32)

	acc := mockAccount{addr, coins, pubKey, accNumber, sequence}
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	cdc.RegisterConcrete(&mockAccount{}, "cosmos-sdk/mockAccount", nil)
	ctx = ctx.WithCodec(cdc)

	// setup expected results
	jsonNoIndent, err := ctx.Codec.MarshalJSON(acc)
	require.Nil(t, err)
	jsonWithIndent, err := ctx.Codec.MarshalJSONIndent(acc, "", "  ")
	require.Nil(t, err)
	respNoIndent := NewResponseWithHeight(height, jsonNoIndent)
	respWithIndent := NewResponseWithHeight(height, jsonWithIndent)
	expectedNoIndent, err := ctx.Codec.MarshalJSON(respNoIndent)
	require.Nil(t, err)
	expectedWithIndent, err := ctx.Codec.MarshalJSONIndent(respWithIndent, "", "  ")
	require.Nil(t, err)

	// check that negative height writes an error
	w := httptest.NewRecorder()
	ctx = ctx.WithHeight(-1)
	PostProcessResponse(w, ctx, acc)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	// check that height returns expected response
	ctx = ctx.WithHeight(height)
	runPostProcessResponse(t, ctx, acc, expectedNoIndent, false)
	// check height with indent
	runPostProcessResponse(t, ctx, acc, expectedWithIndent, true)
}

// asserts that ResponseRecorder returns the expected code and body
// runs PostProcessResponse on the objects regular interface and on
// the marshalled struct.
func runPostProcessResponse(t *testing.T, ctx context.CLIContext, obj interface{}, expectedBody []byte, indent bool) {
	if indent {
		ctx.Indent = indent
	}

	// test using regular struct
	w := httptest.NewRecorder()
	PostProcessResponse(w, ctx, obj)
	require.Equal(t, http.StatusOK, w.Code, w.Body)
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	require.Equal(t, expectedBody, body)

	var marshalled []byte
	if indent {
		marshalled, err = ctx.Codec.MarshalJSONIndent(obj, "", "  ")
	} else {
		marshalled, err = ctx.Codec.MarshalJSON(obj)
	}
	require.Nil(t, err)

	// test using marshalled struct
	w = httptest.NewRecorder()
	PostProcessResponse(w, ctx, marshalled)
	require.Equal(t, http.StatusOK, w.Code, w.Body)
	resp = w.Result()
	body, err = ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	require.Equal(t, expectedBody, body)
}

func mustNewRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	err = req.ParseForm()
	require.NoError(t, err)
	return req
}
