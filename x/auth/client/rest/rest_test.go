// +build norace

package rest_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	rest2 "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func mkTx() legacytx.StdTx {
	// NOTE: this uses StdTx explicitly, don't migrate it!
	return legacytx.StdTx{
		Msgs: []sdk.Msg{&types.MsgSend{}},
		Fee: legacytx.StdFee{
			Amount: sdk.Coins{sdk.NewInt64Coin("foo", 10)},
			Gas:    10000,
		},
		Memo: "FOOBAR",
	}
}

func (s *IntegrationTestSuite) TestEncodeDecode() {
	var require = s.Require()
	val := s.network.Validators[0]
	stdTx := mkTx()

	// NOTE: this uses amino explicitly, don't migrate it!
	cdc := val.ClientCtx.LegacyAmino

	bz, err := cdc.MarshalJSON(stdTx)
	require.NoError(err)

	res, err := rest.PostRequest(fmt.Sprintf("%s/txs/encode", val.APIAddress), "application/json", bz)
	require.NoError(err)

	var encodeResp rest2.EncodeResp
	err = cdc.UnmarshalJSON(res, &encodeResp)
	require.NoError(err)

	bz, err = cdc.MarshalJSON(rest2.DecodeReq{Tx: encodeResp.Tx})
	require.NoError(err)

	res, err = rest.PostRequest(fmt.Sprintf("%s/txs/decode", val.APIAddress), "application/json", bz)
	require.NoError(err)

	var respWithHeight rest.ResponseWithHeight
	err = cdc.UnmarshalJSON(res, &respWithHeight)
	require.NoError(err)
	var decodeResp rest2.DecodeResp
	err = cdc.UnmarshalJSON(respWithHeight.Result, &decodeResp)
	require.NoError(err)
	require.Equal(stdTx, legacytx.StdTx(decodeResp))
}

func (s *IntegrationTestSuite) TestBroadcastTxRequest() {
	stdTx := mkTx()

	// we just test with async mode because this tx will fail - all we care about is that it got encoded and broadcast correctly
	res, err := s.broadcastReq(stdTx, "async")
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	// NOTE: this uses amino explicitly, don't migrate it!
	s.Require().NoError(s.cfg.LegacyAmino.UnmarshalJSON(res, &txRes))
	// we just check for a non-empty TxHash here, the actual hash will depend on the underlying tx configuration
	s.Require().NotEmpty(txRes.TxHash)
}

func (s *IntegrationTestSuite) broadcastReq(stdTx legacytx.StdTx, mode string) ([]byte, error) {
	val := s.network.Validators[0]

	// NOTE: this uses amino explicitly, don't migrate it!
	cdc := val.ClientCtx.LegacyAmino
	req := rest2.BroadcastReq{
		Tx:   stdTx,
		Mode: mode,
	}
	bz, err := cdc.MarshalJSON(req)
	s.Require().NoError(err)

	return rest.PostRequest(fmt.Sprintf("%s/txs", val.APIAddress), "application/json", bz)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
