// Package rest provides HTTP types and primitives for REST
// requests validation and responses handling.
package rest

import (
	//	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	//	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

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
	expectedJSONWithHeight := []byte(`{"height":194423","type":"cosmos-sdk/BaseAccount","value":{"address":"cosmos1gc72g4guwg2efa03xgkx3u6ft6t39xqzelfx04","coins":[{"denom":"atom","amount":"100"},{"denom":"tree","amount":"125"}],"public_key":{"type":"tendermint/PubKeySecp256k1","value":"A0tS6HI8Goq1lXEK2+g2nT5Hq3qteBVtL0va9kn9BB++"},"account_number":"104","sequence":"32"}}`)

	// setup
	var w http.ResponseWriter

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := types.AccAddress(pubKey.Address())
	coins := types.NewCoins(types.NewCoin("atom", types.NewInt(100)), types.NewCoin("tree", types.NewInt(125)))
	//height := int64(194423)
	accNumber := uint64(104)
	sequence := uint64(32)

	acc := authtypes.NewBaseAccount(addr, coins, pubKey, accNumber, sequence)
	cdc := codec.New()
	authtypes.RegisterBaseAccount(cdc)

	// expected json response with zero height
	expectedJSONNoHeight, err := cdc.MarshalJSON(acc)
	require.Nil(t, err)

	ProcessPostResponse(w)

	/*
		To test
		- no height marshaled struct returns same thing
		- not marhsaled no height returns expected no height
		- height marshaled struct returns expected with heigh
		- height no marshaled returns exected with height
		- all the above with indent

			expectedOutput, err := cdc.MarshalJSONIndent(expectedMockAcc, "", " ")
			require.Nil(t, err)
			fmt.Printf("%s\n", expectedOutput)
			fmt.Printf("%s\n", actualOutput)
			//	require.Equal(t, expectedOutput, actualOutput)

			// test that JSONMarshal returns equivalent output
			actualOutput, err = cdc.MarshalJSON(mockAcc)
			require.Nil(t, err)
			m := make(map[string]interface{})
			err = json.Unmarshal(actualOutput, &m)
			require.Nil(t, err)
			fmt.Printf("%s\n", m)
			m["height"] = height
			actualOutput, err = json.MarshalIndent(m, "", " ")
			fmt.Printf("%s\n", actualOutput)
			var i mockAccount
			err = cdc.UnmarshalJSON(actualOutput, &i)
			require.Nil(t, err)
			fmt.Printf("%v\n", i)
	*/
}

func mustNewRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	err = req.ParseForm()
	require.NoError(t, err)
	return req
}
