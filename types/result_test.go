package types_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestParseABCILog(t *testing.T) {
	t.Parallel()
	logs := `[{"log":"","msg_index":1,"success":true}]`

	res, err := sdk.ParseABCILogs(logs)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, res[0].Log, "")
	require.Equal(t, res[0].MsgIndex, uint32(1))
}

func TestABCIMessageLog(t *testing.T) {
	t.Parallel()
	cdc := codec.NewLegacyAmino()
	events := sdk.Events{sdk.NewEvent("transfer", sdk.NewAttribute("sender", "foo"))}
	msgLog := sdk.NewABCIMessageLog(0, "", events)

	msgLogs := sdk.ABCIMessageLogs{msgLog}
	bz, err := cdc.MarshalJSON(msgLogs)
	require.NoError(t, err)
	require.Equal(t, string(bz), msgLogs.String())
}

func TestNewSearchTxsResult(t *testing.T) {
	t.Parallel()
	got := sdk.NewSearchTxsResult(150, 20, 2, 20, []*sdk.TxResponse{})
	require.Equal(t, &sdk.SearchTxsResult{
		TotalCount: 150,
		Count:      20,
		PageNumber: 2,
		PageTotal:  8,
		Limit:      20,
		Txs:        []*sdk.TxResponse{},
	}, got)
}

/*
	Codespace: res.TxResult.Codespace,
	Code:      res.TxResult.Code,
	Data:      strings.ToUpper(hex.EncodeToString(res.TxResult.Data)),
	RawLog:    res.TxResult.Log,
	Logs:      parsedLogs,
	Info:      res.TxResult.Info,
	GasWanted: res.TxResult.GasWanted,
	GasUsed:   res.TxResult.GasUsed,
	Tx:        tx,
	Timestamp: timestamp,
*/

func TestResponseResultTx(t *testing.T) {
	t.Parallel()
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
	require.NoError(t, err)
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

	require.Equal(t, want, sdk.NewResponseResultTx(resultTx, nil, "timestamp"))
	require.Equal(t, (*sdk.TxResponse)(nil), sdk.NewResponseResultTx(nil, nil, "timestamp"))
	require.Equal(t, `Response:
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
	require.True(t, sdk.TxResponse{}.Empty())
	require.False(t, want.Empty())

	resultBroadcastTx := &ctypes.ResultBroadcastTx{
		Code:      1,
		Codespace: "codespace",
		Data:      []byte("data"),
		Log:       `[]`,
		Hash:      bytes.HexBytes([]byte("test")),
	}
	require.Equal(t, &sdk.TxResponse{
		Code:      1,
		Codespace: "codespace",
		Data:      "64617461",
		RawLog:    `[]`,
		Logs:      logs,
		TxHash:    "74657374",
	}, sdk.NewResponseFormatBroadcastTx(resultBroadcastTx))

	require.Equal(t, (*sdk.TxResponse)(nil), sdk.NewResponseFormatBroadcastTx(nil))
}

func TestResponseFormatBroadcastTxCommit(t *testing.T) {
	// test nil
	require.Equal(t, (*sdk.TxResponse)(nil), sdk.NewResponseFormatBroadcastTxCommit(nil))

	logs, err := sdk.ParseABCILogs(`[]`)
	require.NoError(t, err)

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
	require.Equal(t, want, sdk.NewResponseFormatBroadcastTxCommit(checkTxResult))
	require.Equal(t, want, sdk.NewResponseFormatBroadcastTxCommit(deliverTxResult))
}
