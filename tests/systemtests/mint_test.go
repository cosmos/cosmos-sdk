//go:build system_test

package systemtests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	systest "cosmossdk.io/systemtests"
)

func TestMintQueries(t *testing.T) {
	// scenario: test mint grpc queries
	// given a running chain

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	systest.Sut.ModifyGenesisJSON(t,
		func(genesis []byte) []byte {
			state, err := sjson.Set(string(genesis), "app_state.mint.minter.inflation", "1.00")
			require.NoError(t, err)
			return []byte(state)
		},
		func(genesis []byte) []byte {
			state, err := sjson.Set(string(genesis), "app_state.mint.params.inflation_max", "1.00")
			require.NoError(t, err)
			return []byte(state)
		},
	)

	systest.Sut.StartChain(t)

	systest.Sut.AwaitNextBlock(t)

	baseurl := systest.Sut.APIAddress()
	blockHeightHeader := "x-cosmos-block-height"
	queryAtHeight := "1"
	paramsResp := `{"params":{"mint_denom":"stake","inflation_rate_change":"0.130000000000000000","inflation_max":"1.000000000000000000","inflation_min":"0.000000000000000000","goal_bonded":"0.670000000000000000","blocks_per_year":"6311520","max_supply":"0"}}`
	inflationResp := `{"inflation":"1.000000000000000000"}`
	annualProvisionsResp := `{"annual_provisions":"2000000000.000000000000000000"}`

	testCases := []struct {
		name    string
		url     string
		headers map[string]string
		expOut  string
	}{
		{
			"gRPC request params",
			fmt.Sprintf("%s/cosmos/mint/v1beta1/params", baseurl),
			map[string]string{},
			paramsResp,
		},
		{
			"gRPC request inflation",
			fmt.Sprintf("%s/cosmos/mint/v1beta1/inflation", baseurl),
			map[string]string{},
			inflationResp,
		},
		{
			"gRPC request annual provisions",
			fmt.Sprintf("%s/cosmos/mint/v1beta1/annual_provisions", baseurl),
			map[string]string{
				blockHeightHeader: queryAtHeight,
			},
			annualProvisionsResp,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: remove below check once grpc gateway is implemented in v2
			if systest.IsV2() {
				return
			}
			resp := systest.GetRequestWithHeaders(t, tc.url, tc.headers, http.StatusOK)
			require.JSONEq(t, tc.expOut, string(resp))
		})
	}

	// test cli queries
	rsp := cli.CustomQuery("q", "mint", "params")
	require.JSONEq(t, paramsResp, rsp)

	rsp = cli.CustomQuery("q", "mint", "inflation")
	require.JSONEq(t, inflationResp, rsp)

	rsp = cli.CustomQuery("q", "mint", "annual-provisions", "--height="+queryAtHeight)
	require.JSONEq(t, annualProvisionsResp, rsp)
}
