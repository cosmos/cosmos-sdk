package rest_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

func TestBaseReq_Sanitize(t *testing.T) {
	t.Parallel()
	sanitized := rest.BaseReq{ChainID: "   test",
		Memo:          "memo     ",
		From:          " cosmos1cq0sxam6x4l0sv9yz3a2vlqhdhvt2k6jtgcse0 ",
		Gas:           " ",
		GasAdjustment: "  0.3",
	}.Sanitize()
	require.Equal(t, rest.BaseReq{ChainID: "test",
		Memo:          "memo",
		From:          "cosmos1cq0sxam6x4l0sv9yz3a2vlqhdhvt2k6jtgcse0",
		Gas:           "",
		GasAdjustment: "0.3",
	}, sanitized)
}

func TestBaseReq_ValidateBasic(t *testing.T) {
	fromAddr := "cosmos1cq0sxam6x4l0sv9yz3a2vlqhdhvt2k6jtgcse0"
	tenstakes, err := types.ParseCoinsNormalized("10stake")
	require.NoError(t, err)
	onestake, err := types.ParseDecCoins("1.0stake")
	require.NoError(t, err)

	req1 := rest.NewBaseReq(
		fromAddr, "", "nonempty", "", "", 0, 0, tenstakes, nil, false,
	)
	req2 := rest.NewBaseReq(
		"", "", "nonempty", "", "", 0, 0, tenstakes, nil, false,
	)
	req3 := rest.NewBaseReq(
		fromAddr, "", "", "", "", 0, 0, tenstakes, nil, false,
	)
	req4 := rest.NewBaseReq(
		fromAddr, "", "nonempty", "", "", 0, 0, tenstakes, onestake, false,
	)
	req5 := rest.NewBaseReq(
		fromAddr, "", "nonempty", "", "", 0, 0, types.Coins{}, types.DecCoins{}, false,
	)

	tests := []struct {
		name string
		req  rest.BaseReq
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.req.ValidateBasic(tt.w))
		})
	}
}

func TestParseHTTPArgs(t *testing.T) {
	t.Parallel()
	req0 := mustNewRequest(t, "", "/", nil)
	req1 := mustNewRequest(t, "", "/?limit=5", nil)
	req2 := mustNewRequest(t, "", "/?page=5", nil)
	req3 := mustNewRequest(t, "", "/?page=5&limit=5", nil)

	reqE1 := mustNewRequest(t, "", "/?page=-1", nil)
	reqE2 := mustNewRequest(t, "", "/?limit=-1", nil)
	req4 := mustNewRequest(t, "", "/?foo=faa", nil)

	reqTxH := mustNewRequest(t, "", "/?tx.minheight=12&tx.maxheight=14", nil)

	tests := []struct {
		name  string
		req   *http.Request
		w     http.ResponseWriter
		tags  []string
		page  int
		limit int
		err   bool
	}{
		{"no params", req0, httptest.NewRecorder(), []string{}, rest.DefaultPage, rest.DefaultLimit, false},
		{"Limit", req1, httptest.NewRecorder(), []string{}, rest.DefaultPage, 5, false},
		{"Page", req2, httptest.NewRecorder(), []string{}, 5, rest.DefaultLimit, false},
		{"Page and limit", req3, httptest.NewRecorder(), []string{}, 5, 5, false},

		{"error page 0", reqE1, httptest.NewRecorder(), []string{}, rest.DefaultPage, rest.DefaultLimit, true},
		{"error limit 0", reqE2, httptest.NewRecorder(), []string{}, rest.DefaultPage, rest.DefaultLimit, true},

		{"tags", req4, httptest.NewRecorder(), []string{"foo='faa'"}, rest.DefaultPage, rest.DefaultLimit, false},
		{"tags", reqTxH, httptest.NewRecorder(), []string{"tx.height<=14", "tx.height>=12"}, rest.DefaultPage, rest.DefaultLimit, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tags, page, limit, err := rest.ParseHTTPArgs(tt.req)

			sort.Strings(tags)

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
	t.Parallel()
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
		clientCtx      client.Context
		expectedHeight int64
		expectedOk     bool
	}{
		{"no height", req0, httptest.NewRecorder(), client.Context{}, emptyHeight, true},
		{"height", req1, httptest.NewRecorder(), client.Context{}, height, true},
		{"invalid height", req2, httptest.NewRecorder(), client.Context{}, emptyHeight, false},
		{"negative height", req3, httptest.NewRecorder(), client.Context{}, emptyHeight, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(tt.w, tt.clientCtx, tt.req)
			if tt.expectedOk {
				require.True(t, ok)
				require.Equal(t, tt.expectedHeight, clientCtx.Height)
			} else {
				require.False(t, ok)
				require.Empty(t, tt.expectedHeight, clientCtx.Height)
			}
		})
	}
}

func TestProcessPostResponse(t *testing.T) {
	// mock account
	// PubKey field ensures amino encoding is used first since standard
	// JSON encoding will panic on cryptotypes.PubKey

	t.Parallel()
	type mockAccount struct {
		Address       types.AccAddress   `json:"address"`
		Coins         types.Coins        `json:"coins"`
		PubKey        cryptotypes.PubKey `json:"public_key"`
		AccountNumber uint64             `json:"account_number"`
		Sequence      uint64             `json:"sequence"`
	}

	// setup
	viper.Set(flags.FlagOffline, true)
	ctx := client.Context{}
	height := int64(194423)

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := types.AccAddress(pubKey.Address())
	coins := types.NewCoins(types.NewCoin("atom", types.NewInt(100)), types.NewCoin("tree", types.NewInt(125)))
	accNumber := uint64(104)
	sequence := uint64(32)

	acc := mockAccount{addr, coins, pubKey, accNumber, sequence}
	cdc := codec.NewLegacyAmino()
	cryptocodec.RegisterCrypto(cdc)
	cdc.RegisterConcrete(&mockAccount{}, "cosmos-sdk/mockAccount", nil)
	ctx = ctx.WithLegacyAmino(cdc)

	// setup expected results
	jsonNoIndent, err := ctx.LegacyAmino.MarshalJSON(acc)
	require.Nil(t, err)

	respNoIndent := rest.NewResponseWithHeight(height, jsonNoIndent)
	expectedNoIndent, err := ctx.LegacyAmino.MarshalJSON(respNoIndent)
	require.Nil(t, err)

	// check that negative height writes an error
	w := httptest.NewRecorder()
	ctx = ctx.WithHeight(-1)
	rest.PostProcessResponse(w, ctx, acc)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	// check that height returns expected response
	ctx = ctx.WithHeight(height)
	runPostProcessResponse(t, ctx, acc, expectedNoIndent)
}

func TestReadRESTReq(t *testing.T) {
	t.Parallel()
	reqBody := ioutil.NopCloser(strings.NewReader(`{"chain_id":"alessio","memo":"text"}`))
	req := &http.Request{Body: reqBody}
	w := httptest.NewRecorder()
	var br rest.BaseReq

	// test OK
	rest.ReadRESTReq(w, req, codec.NewLegacyAmino(), &br)
	res := w.Result() //nolint:bodyclose
	t.Cleanup(func() { res.Body.Close() })
	require.Equal(t, rest.BaseReq{ChainID: "alessio", Memo: "text"}, br)
	require.Equal(t, http.StatusOK, res.StatusCode)

	// test non valid JSON
	reqBody = ioutil.NopCloser(strings.NewReader(`MALFORMED`))
	req = &http.Request{Body: reqBody}
	br = rest.BaseReq{}
	w = httptest.NewRecorder()
	rest.ReadRESTReq(w, req, codec.NewLegacyAmino(), &br)
	require.Equal(t, br, br)
	res = w.Result() //nolint:bodyclose
	t.Cleanup(func() { res.Body.Close() })
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestWriteSimulationResponse(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	rest.WriteSimulationResponse(w, codec.NewLegacyAmino(), 10)
	res := w.Result() //nolint:bodyclose
	t.Cleanup(func() { res.Body.Close() })
	require.Equal(t, http.StatusOK, res.StatusCode)
	bs, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	t.Cleanup(func() { res.Body.Close() })
	require.Equal(t, `{"gas_estimate":"10"}`, string(bs))
}

func TestParseUint64OrReturnBadRequest(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	_, ok := rest.ParseUint64OrReturnBadRequest(w, "100")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, w.Result().StatusCode) //nolint:bodyclose

	w = httptest.NewRecorder()
	_, ok = rest.ParseUint64OrReturnBadRequest(w, "-100")
	require.False(t, ok)
	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode) //nolint:bodyclose
}

func TestParseFloat64OrReturnBadRequest(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	_, ok := rest.ParseFloat64OrReturnBadRequest(w, "100", 0)
	require.True(t, ok)
	require.Equal(t, http.StatusOK, w.Result().StatusCode) //nolint:bodyclose

	w = httptest.NewRecorder()
	_, ok = rest.ParseFloat64OrReturnBadRequest(w, "bad request", 0)
	require.False(t, ok)
	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode) //nolint:bodyclose

	w = httptest.NewRecorder()
	ret, ok := rest.ParseFloat64OrReturnBadRequest(w, "", 9.0)
	require.Equal(t, float64(9), ret)
	require.True(t, ok)
	require.Equal(t, http.StatusOK, w.Result().StatusCode) //nolint:bodyclose
}

func TestParseQueryParamBool(t *testing.T) {
	req := httptest.NewRequest("GET", "/target?boolean=true", nil)
	require.True(t, rest.ParseQueryParamBool(req, "boolean"))
	require.False(t, rest.ParseQueryParamBool(req, "nokey"))
	req = httptest.NewRequest("GET", "/target?boolean=false", nil)
	require.False(t, rest.ParseQueryParamBool(req, "boolean"))
	require.False(t, rest.ParseQueryParamBool(req, ""))
}

func TestPostProcessResponseBare(t *testing.T) {
	t.Parallel()

	encodingConfig := simappparams.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino) // amino used intentionally here
	// write bytes
	w := httptest.NewRecorder()
	bs := []byte("text string")

	rest.PostProcessResponseBare(w, clientCtx, bs)

	res := w.Result() //nolint:bodyclose
	require.Equal(t, http.StatusOK, res.StatusCode)

	got, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	t.Cleanup(func() { res.Body.Close() })
	require.Equal(t, "text string", string(got))

	// write struct and indent response
	w = httptest.NewRecorder()
	data := struct {
		X int    `json:"x"`
		S string `json:"s"`
	}{X: 10, S: "test"}

	rest.PostProcessResponseBare(w, clientCtx, data)

	res = w.Result() //nolint:bodyclose
	require.Equal(t, http.StatusOK, res.StatusCode)

	got, err = ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	t.Cleanup(func() { res.Body.Close() })
	require.Equal(t, "{\"x\":\"10\",\"s\":\"test\"}", string(got))

	// write struct, don't indent response
	w = httptest.NewRecorder()
	data = struct {
		X int    `json:"x"`
		S string `json:"s"`
	}{X: 10, S: "test"}

	rest.PostProcessResponseBare(w, clientCtx, data)

	res = w.Result() //nolint:bodyclose
	require.Equal(t, http.StatusOK, res.StatusCode)

	got, err = ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	t.Cleanup(func() { res.Body.Close() })
	require.Equal(t, `{"x":"10","s":"test"}`, string(got))

	// test marshalling failure
	w = httptest.NewRecorder()
	data2 := badJSONMarshaller{}

	rest.PostProcessResponseBare(w, clientCtx, data2)

	res = w.Result() //nolint:bodyclose
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	got, err = ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	t.Cleanup(func() { res.Body.Close() })
	require.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
	require.Equal(t, `{"error":"couldn't marshal"}`, string(got))
}

type badJSONMarshaller struct{}

func (badJSONMarshaller) MarshalJSON() ([]byte, error) {
	return nil, errors.New("couldn't marshal")
}

// asserts that ResponseRecorder returns the expected code and body
// runs PostProcessResponse on the objects regular interface and on
// the marshalled struct.
func runPostProcessResponse(t *testing.T, ctx client.Context, obj interface{}, expectedBody []byte) {
	// test using regular struct
	w := httptest.NewRecorder()

	rest.PostProcessResponse(w, ctx, obj)
	require.Equal(t, http.StatusOK, w.Code, w.Body)

	resp := w.Result() //nolint:bodyclose
	t.Cleanup(func() { resp.Body.Close() })

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	require.Equal(t, expectedBody, body)

	marshalled, err := ctx.LegacyAmino.MarshalJSON(obj)
	require.NoError(t, err)

	// test using marshalled struct
	w = httptest.NewRecorder()
	rest.PostProcessResponse(w, ctx, marshalled)

	require.Equal(t, http.StatusOK, w.Code, w.Body)
	resp = w.Result() //nolint:bodyclose

	t.Cleanup(func() { resp.Body.Close() })
	body, err = ioutil.ReadAll(resp.Body)

	require.Nil(t, err)
	require.Equal(t, string(expectedBody), string(body))
}

func mustNewRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	err = req.ParseForm()
	require.NoError(t, err)
	return req
}

func TestCheckErrors(t *testing.T) {
	t.Parallel()
	err := errors.New("ERROR")
	tests := []struct {
		name       string
		checkerFn  func(w http.ResponseWriter, err error) bool
		error      error
		wantErr    bool
		wantString string
		wantStatus int
	}{
		{"500", rest.CheckInternalServerError, err, true, `{"error":"ERROR"}`, http.StatusInternalServerError},
		{"500 (no error)", rest.CheckInternalServerError, nil, false, ``, http.StatusInternalServerError},
		{"400", rest.CheckBadRequestError, err, true, `{"error":"ERROR"}`, http.StatusBadRequest},
		{"400 (no error)", rest.CheckBadRequestError, nil, false, ``, http.StatusBadRequest},
		{"404", rest.CheckNotFoundError, err, true, `{"error":"ERROR"}`, http.StatusNotFound},
		{"404 (no error)", rest.CheckNotFoundError, nil, false, ``, http.StatusNotFound},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			require.Equal(t, tt.wantErr, tt.checkerFn(w, tt.error))
			if tt.wantErr {
				require.Equal(t, w.Body.String(), tt.wantString)
				require.Equal(t, w.Code, tt.wantStatus)
			}
		})
	}
}
