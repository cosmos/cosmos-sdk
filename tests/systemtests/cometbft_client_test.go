//go:build system_test

package systemtests

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
)

func TestQueryNodeInfo(t *testing.T) {
	baseurl := fmt.Sprintf("http://localhost:%d", apiPortStart)
	sut.ResetChain(t)
	sut.StartChain(t)

	qc := cmtservice.NewServiceClient(sut.RPCClient(t))
	res, err := qc.GetNodeInfo(context.Background(), &cmtservice.GetNodeInfoRequest{})
	assert.NoError(t, err)

	v := NewCLIWrapper(t, sut, true).Version()
	assert.Equal(t, res.ApplicationVersion.Version, v)

	// TODO: we should be adding a way to distinguish a v2. Eventually we should skip some v2 system depending on the consensus engine we want to test
	restRes := GetRequest(t, mustV(url.JoinPath(baseurl, "/cosmos/base/tendermint/v1beta1/node_info")))
	assert.NoError(t, err)
	assert.Equal(t, gjson.GetBytes(restRes, "application_version.version").String(), res.ApplicationVersion.Version)
}

func TestQuerySyncing(t *testing.T) {
	baseurl := fmt.Sprintf("http://localhost:%d", apiPortStart)
	sut.ResetChain(t)
	sut.StartChain(t)

	qc := cmtservice.NewServiceClient(sut.RPCClient(t))
	res, err := qc.GetSyncing(context.Background(), &cmtservice.GetSyncingRequest{})
	assert.NoError(t, err)

	restRes := GetRequest(t, mustV(url.JoinPath(baseurl, "/cosmos/base/tendermint/v1beta1/syncing")))
	assert.Equal(t, gjson.GetBytes(restRes, "syncing").Bool(), res.Syncing)
}

func TestQueryLatestBlock(t *testing.T) {
	baseurl := fmt.Sprintf("http://localhost:%d", apiPortStart)
	sut.ResetChain(t)
	sut.StartChain(t)

	qc := cmtservice.NewServiceClient(sut.RPCClient(t))
	res, err := qc.GetLatestBlock(context.Background(), &cmtservice.GetLatestBlockRequest{})
	assert.NoError(t, err)
	assert.Contains(t, res.SdkBlock.Header.ProposerAddress, "cosmosvalcons")

	_ = GetRequest(t, mustV(url.JoinPath(baseurl, "/cosmos/base/tendermint/v1beta1/blocks/latest")))
}

func TestQueryBlockByHeight(t *testing.T) {
	baseurl := fmt.Sprintf("http://localhost:%d", apiPortStart)
	sut.ResetChain(t)
	sut.StartChain(t)

	sut.AwaitNBlocks(t, 2, time.Second*25)

	qc := cmtservice.NewServiceClient(sut.RPCClient(t))
	res, err := qc.GetBlockByHeight(context.Background(), &cmtservice.GetBlockByHeightRequest{Height: 2})
	assert.NoError(t, err)
	assert.Equal(t, res.SdkBlock.Header.Height, int64(2))
	assert.Contains(t, res.SdkBlock.Header.ProposerAddress, "cosmosvalcons")

	restRes := GetRequest(t, mustV(url.JoinPath(baseurl, "/cosmos/base/tendermint/v1beta1/blocks/2")))
	assert.Equal(t, gjson.GetBytes(restRes, "sdk_block.header.height").Int(), int64(2))
	assert.Contains(t, gjson.GetBytes(restRes, "sdk_block.header.proposer_address").String(), "cosmosvalcons")
}

func TestQueryLatestValidatorSet(t *testing.T) {
	if sut.NodesCount() < 2 {
		t.Skip("not enough nodes")
		return
	}
	baseurl := fmt.Sprintf("http://localhost:%d", apiPortStart)
	sut.ResetChain(t)
	sut.StartChain(t)

	vals := sut.RPCClient(t).Validators()

	qc := cmtservice.NewServiceClient(sut.RPCClient(t))
	res, err := qc.GetLatestValidatorSet(context.Background(), &cmtservice.GetLatestValidatorSetRequest{
		Pagination: nil,
	})
	assert.NoError(t, err)
	assert.Equal(t, len(res.Validators), len(vals))

	// with pagination
	res, err = qc.GetLatestValidatorSet(context.Background(), &cmtservice.GetLatestValidatorSetRequest{Pagination: &qtypes.PageRequest{
		Offset: 0,
		Limit:  2,
	}})
	assert.NoError(t, err)
	assert.Equal(t, len(res.Validators), 2)

	restRes := GetRequest(t, fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest?pagination.offset=%d&pagination.limit=%d", baseurl, 0, 2))
	assert.Equal(t, len(gjson.GetBytes(restRes, "validators").Array()), 2)
}

func TestLatestValidatorSet(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	vals := sut.RPCClient(t).Validators()

	qc := cmtservice.NewServiceClient(sut.RPCClient(t))
	testCases := []struct {
		name      string
		req       *cmtservice.GetLatestValidatorSetRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "cannot be nil"},
		{"no pagination", &cmtservice.GetLatestValidatorSetRequest{}, false, ""},
		{"with pagination", &cmtservice.GetLatestValidatorSetRequest{Pagination: &qtypes.PageRequest{Offset: 0, Limit: uint64(len(vals))}}, false, ""},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := qc.GetLatestValidatorSet(context.Background(), tc.req)
			if tc.expErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(res.Validators), len(vals))
				content, ok := res.Validators[0].PubKey.GetCachedValue().(cryptotypes.PubKey)
				assert.True(t, ok)
				assert.Equal(t, content.Address(), vals[0].PubKey.Address())
			}
		})
	}
}

func TestLatestValidatorSet_GRPCGateway(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	baseurl := fmt.Sprintf("http://localhost:%d", apiPortStart)

	vals := sut.RPCClient(t).Validators()

	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{"no pagination", "/cosmos/base/tendermint/v1beta1/validatorsets/latest", false, ""},
		{"pagination invalid fields", "/cosmos/base/tendermint/v1beta1/validatorsets/latest?pagination.offset=-1&pagination.limit=-2", true, "strconv.ParseUint"},
		{"with pagination", "/cosmos/base/tendermint/v1beta1/validatorsets/latest?pagination.offset=0&pagination.limit=2", false, ""},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expErr {
				rsp := GetRequestWithHeaders(t, baseurl+tc.url, nil, http.StatusBadRequest)
				errMsg := gjson.GetBytes(rsp, "message").String()
				assert.Contains(t, errMsg, tc.expErrMsg)
				return
			}
			rsp := GetRequest(t, baseurl+tc.url)
			assert.Equal(t, len(vals), int(gjson.GetBytes(rsp, "pagination.total").Int()))
		})
	}
}

func TestValidatorSetByHeight(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	qc := cmtservice.NewServiceClient(sut.RPCClient(t))
	vals := sut.RPCClient(t).Validators()

	testCases := []struct {
		name      string
		req       *cmtservice.GetValidatorSetByHeightRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &cmtservice.GetValidatorSetByHeightRequest{}, true, "height must be greater than 0"},
		{"no pagination", &cmtservice.GetValidatorSetByHeightRequest{Height: 1}, false, ""},
		{"with pagination", &cmtservice.GetValidatorSetByHeightRequest{Height: 1, Pagination: &qtypes.PageRequest{Offset: 0, Limit: uint64(len(vals))}}, false, ""},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := qc.GetValidatorSetByHeight(context.Background(), tc.req)
			if tc.expErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(res.Validators), len(vals))
			}
		})
	}
}

func TestValidatorSetByHeight_GRPCRestGateway(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	vals := sut.RPCClient(t).Validators()

	baseurl := sut.APIAddress()
	block := sut.AwaitNextBlock(t, time.Second*3)
	testCases := []struct {
		name        string
		url         string
		expErr      bool
		expErrMsg   string
		expHttpCode int
	}{
		{"invalid height", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d", baseurl, -1), true, "height must be greater than 0", http.StatusInternalServerError},
		{"no pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d", baseurl, block), false, "", http.StatusOK},
		{"pagination invalid fields", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d?pagination.offset=-1&pagination.limit=-2", baseurl, block), true, "strconv.ParseUint", http.StatusBadRequest},
		{"with pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d?pagination.limit=2", baseurl, 1), false, "", http.StatusOK},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rsp := GetRequestWithHeaders(t, tc.url, nil, tc.expHttpCode)
			if tc.expErr {
				errMsg := gjson.GetBytes(rsp, "message").String()
				assert.Contains(t, errMsg, tc.expErrMsg)
				return
			}
			assert.Equal(t, len(vals), int(gjson.GetBytes(rsp, "pagination.total").Int()))
		})
	}
}

func TestABCIQuery(t *testing.T) {
	sut.StartChain(t)

	qc := cmtservice.NewServiceClient(sut.RPCClient(t))
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	testCases := []struct {
		name         string
		req          *cmtservice.ABCIQueryRequest
		expectErr    bool
		expectedCode uint32
		validQuery   bool
	}{
		{
			name: "valid request with proof",
			req: &cmtservice.ABCIQueryRequest{
				Path:  "/store/gov/key",
				Data:  []byte{0x03},
				Prove: true,
			},
			validQuery: true,
		},
		{
			name: "valid request without proof",
			req: &cmtservice.ABCIQueryRequest{
				Path:  "/store/gov/key",
				Data:  []byte{0x03},
				Prove: false,
			},
			validQuery: true,
		},
		{
			name: "request with invalid path",
			req: &cmtservice.ABCIQueryRequest{
				Path: "/foo/bar",
				Data: []byte{0x03},
			},
			expectErr: true,
		},
		{
			name: "request with invalid path recursive",
			req: &cmtservice.ABCIQueryRequest{
				Path: "/cosmos.base.tendermint.v1beta1.Service/ABCIQuery",
				Data: cdc.MustMarshal(&cmtservice.ABCIQueryRequest{
					Path: "/cosmos.base.tendermint.v1beta1.Service/ABCIQuery",
				}),
			},
			expectErr: true,
		},
		{
			name: "request with invalid broadcast tx path",
			req: &cmtservice.ABCIQueryRequest{
				Path: "/cosmos.tx.v1beta1.Service/BroadcastTx",
				Data: []byte{0x00},
			},
			expectErr: true,
		},
		{
			name: "request with invalid data",
			req: &cmtservice.ABCIQueryRequest{
				Path: "/store/gov/key",
				Data: []byte{0x0044, 0x00},
			},
			validQuery: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := qc.ABCIQuery(context.Background(), tc.req)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tc.expectedCode, res.Code)
			}

			if tc.validQuery {
				assert.Greater(t, res.Height, int64(0))
				assert.Greater(t, len(res.Key), 0)
				assert.Greater(t, len(res.Value), 0)
			}
		})
	}
}
