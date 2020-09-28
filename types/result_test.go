package types_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type resultTestStuite struct {
	suite.Suite
}

func TestRTestStuite(t *testing.T) {
	suite.Run(t, new(resultTestStuite))
}

func (s *resultTestStuite) TestParseABCILog() {
	logs := `[{"log":"","msg_index":1,"success":true}]`
	res, err := sdk.ParseABCILogs(logs)

	s.Require().NoError(err)
	s.Require().Len(res, 1)
	s.Require().Equal(res[0].Log, "")
	s.Require().Equal(res[0].MsgIndex, uint32(1))
}

func (s *resultTestStuite) TestABCIMessageLog() {
	cdc := codec.NewLegacyAmino()
	events := sdk.Events{sdk.NewEvent("transfer", sdk.NewAttribute("sender", "foo"))}
	msgLog := sdk.NewABCIMessageLog(0, "", events)
	msgLogs := sdk.ABCIMessageLogs{msgLog}
	bz, err := cdc.MarshalJSON(msgLogs)

	s.Require().NoError(err)
	s.Require().Equal(string(bz), msgLogs.String())
}

func (s *resultTestStuite) TestNewSearchTxsResult() {
	got := sdk.NewSearchTxsResult(150, 20, 2, 20, []*sdk.TxResponse{})
	s.Require().Equal(&sdk.SearchTxsResult{
		TotalCount: 150,
		Count:      20,
		PageNumber: 2,
		PageTotal:  8,
		Limit:      20,
		Txs:        []*sdk.TxResponse{},
	}, got)
}

func (s *resultTestStuite) TestResponseResultTx() {
	deliverTxResult := abci.ResponseDeliverTx{
		Codespace: "codespace",
		Code:      1,
		Data:      []byte("data"),
		Log:       `[]`,
		Info:      "info",
		GasWanted: 100,
		GasUsed:   90,
	}
	resultTx := &ctypes.ResultTx{
		Hash:     bytes.HexBytes([]byte("test")),
		Height:   10,
		TxResult: deliverTxResult,
	}
	logs, err := sdk.ParseABCILogs(`[]`)

	s.Require().NoError(err)

	want := &sdk.TxResponse{
		TxHash:    "74657374",
		Height:    10,
		Codespace: "codespace",
		Code:      1,
		Data:      strings.ToUpper(hex.EncodeToString([]byte("data"))),
		RawLog:    `[]`,
		Logs:      logs,
		Info:      "info",
		GasWanted: 100,
		GasUsed:   90,
		Tx:        nil,
		Timestamp: "timestamp",
	}

	s.Require().Equal(want, sdk.NewResponseResultTx(resultTx, nil, "timestamp"))
	s.Require().Equal((*sdk.TxResponse)(nil), sdk.NewResponseResultTx(nil, nil, "timestamp"))
	s.Require().Equal(`Response:
  Height: 10
  TxHash: 74657374
  Code: 1
  Data: 64617461
  Raw Log: []
  Logs: []
  Info: info
  GasWanted: 100
  GasUsed: 90
  Codespace: codespace
  Timestamp: timestamp`, sdk.NewResponseResultTx(resultTx, nil, "timestamp").String())
	s.Require().True(sdk.TxResponse{}.Empty())
	s.Require().False(want.Empty())

	resultBroadcastTx := &ctypes.ResultBroadcastTx{
		Code:      1,
		Codespace: "codespace",
		Data:      []byte("data"),
		Log:       `[]`,
		Hash:      bytes.HexBytes([]byte("test")),
	}

	s.Require().Equal(&sdk.TxResponse{
		Code:      1,
		Codespace: "codespace",
		Data:      "64617461",
		RawLog:    `[]`,
		Logs:      logs,
		TxHash:    "74657374",
	}, sdk.NewResponseFormatBroadcastTx(resultBroadcastTx))
	s.Require().Equal((*sdk.TxResponse)(nil), sdk.NewResponseFormatBroadcastTx(nil))
}

func (s *resultTestStuite) TestResponseFormatBroadcastTxCommit() {
	// test nil
	s.Require().Equal((*sdk.TxResponse)(nil), sdk.NewResponseFormatBroadcastTxCommit(nil))

	logs, err := sdk.ParseABCILogs(`[]`)
	s.Require().NoError(err)

	// test checkTx
	checkTxResult := &ctypes.ResultBroadcastTxCommit{
		Height: 10,
		Hash:   bytes.HexBytes([]byte("test")),
		CheckTx: abci.ResponseCheckTx{
			Code:      90,
			Data:      nil,
			Log:       `[]`,
			Info:      "info",
			GasWanted: 99,
			GasUsed:   100,
			Codespace: "codespace",
		},
	}
	deliverTxResult := &ctypes.ResultBroadcastTxCommit{
		Height: 10,
		Hash:   bytes.HexBytes([]byte("test")),
		DeliverTx: abci.ResponseDeliverTx{
			Code:      90,
			Data:      nil,
			Log:       `[]`,
			Info:      "info",
			GasWanted: 99,
			GasUsed:   100,
			Codespace: "codespace",
		},
	}
	want := &sdk.TxResponse{
		Height:    10,
		TxHash:    "74657374",
		Codespace: "codespace",
		Code:      90,
		Data:      "",
		RawLog:    `[]`,
		Logs:      logs,
		Info:      "info",
		GasWanted: 99,
		GasUsed:   100,
	}

	s.Require().Equal(want, sdk.NewResponseFormatBroadcastTxCommit(checkTxResult))
	s.Require().Equal(want, sdk.NewResponseFormatBroadcastTxCommit(deliverTxResult))
}
