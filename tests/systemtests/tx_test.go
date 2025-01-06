//go:build system_test

package systemtests

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	bankMsgSendEventAction       = "message.action='/cosmos.bank.v1beta1.MsgSend'"
	denom                        = "stake"
	transferAmount         int64 = 1000
)

func TestQueryBySig(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	qc := tx.NewServiceClient(systest.Sut.RPCClient(t))

	// create unsign tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=10stake", "--chain-id=" + cli.ChainID(), "--sign-mode=direct", "--generate-only"}
	unsignedTx := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := systest.StoreTempFile(t, []byte(unsignedTx))

	signedTx := cli.RunCommandWithArgs("tx", "sign", txFile.Name(), "--from="+valAddr, "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home="+systest.Sut.NodeDir(0))
	sig := gjson.Get(signedTx, "signatures.0").String()
	signedTxFile := systest.StoreTempFile(t, []byte(signedTx))

	res := cli.Run("tx", "broadcast", signedTxFile.Name())
	systest.RequireTxSuccess(t, res)

	sigFormatted := fmt.Sprintf("%s.%s='%s'", sdk.EventTypeTx, sdk.AttributeKeySignature, sig)
	resp, err := qc.GetTxsEvent(context.Background(), &tx.GetTxsEventRequest{
		Query:   sigFormatted,
		OrderBy: 0,
		Page:    0,
		Limit:   10,
	})
	require.NoError(t, err)
	require.Len(t, resp.Txs, 1)
	require.Len(t, resp.Txs[0].Signatures, 1)
}

func TestSimulateTx_GRPC(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	qc := tx.NewServiceClient(systest.Sut.RPCClient(t))

	// create unsign tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + cli.ChainID(), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), "--from="+valAddr, "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home="+systest.Sut.NodeDir(0))
	signedTxFile := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "encode", signedTxFile.Name())
	txBz, err := base64.StdEncoding.DecodeString(res)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		req       *tx.SimulateRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.SimulateRequest{}, true, "empty txBytes is not allowed"},
		{"valid request with tx_bytes", &tx.SimulateRequest{TxBytes: txBz}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Broadcast the tx via gRPC via the validator's clientCtx (which goes
			// through Tendermint).
			res, err := qc.Simulate(context.Background(), tc.req)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
				// Check the result and gas used are correct.
				//
				// The 10 events are:
				// - Sending Fee to the pool: coin_spent, coin_received and transfer
				// - tx.* events: tx.fee, tx.acc_seq, tx.signature
				// - Sending Amount to recipient: coin_spent, coin_received and transfer
				// - Msg events: message.module=bank, message.action=/cosmos.bank.v1beta1.MsgSend and message.sender=<val1> (in one message)
				require.Equal(t, 10, len(res.GetResult().GetEvents()))
				require.True(t, res.GetGasInfo().GetGasUsed() > 0) // Gas used sometimes change, just check it's not empty.
			}
		})
	}
}

func TestSimulateTx_GRPCGateway(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	// qc := tx.NewServiceClient(sut.RPCClient(t))
	baseURL := systest.Sut.APIAddress()

	// create unsign tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + cli.ChainID(), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), "--from="+valAddr, "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home="+systest.Sut.NodeDir(0))
	signedTxFile := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "encode", signedTxFile.Name())
	txBz, err := base64.StdEncoding.DecodeString(res)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		req       *tx.SimulateRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.SimulateRequest{}, true, "empty txBytes is not allowed"},
		{"valid request with tx_bytes", &tx.SimulateRequest{TxBytes: txBz}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBz, err := json.Marshal(tc.req)
			require.NoError(t, err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/simulate", baseURL), "application/json", reqBz)
			require.NoError(t, err)
			if tc.expErr {
				require.Contains(t, string(res), tc.expErrMsg)
			} else {
				require.NoError(t, err)
				msgResponses := gjson.Get(string(res), "result.msg_responses").Array()
				require.Equal(t, len(msgResponses), 1)

				events := gjson.Get(string(res), "result.events").Array()
				require.Equal(t, len(events), 10)

				gasUsed := gjson.Get(string(res), "gas_info.gas_used").Int()
				require.True(t, gasUsed > 0)
			}
		})
	}
}

func TestGetTxEvents_GRPC(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	qc := tx.NewServiceClient(systest.Sut.RPCClient(t))
	rsp := cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--note=foobar", "--fees=1stake")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)

	rsp = cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake")
	txResult, found = cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)

	testCases := []struct {
		name      string
		req       *tx.GetTxsEventRequest
		expErr    bool
		expErrMsg string
		expLen    int
	}{
		{
			"nil request",
			nil,
			true,
			"request cannot be nil",
			0,
		},
		{
			"empty request",
			&tx.GetTxsEventRequest{},
			true,
			"query cannot be empty",
			0,
		},
		{
			"request with dummy event",
			&tx.GetTxsEventRequest{Query: "foobar"},
			true,
			"failed to search for txs",
			0,
		},
		{
			"request with order-by",
			&tx.GetTxsEventRequest{
				Query:   bankMsgSendEventAction,
				OrderBy: tx.OrderBy_ORDER_BY_ASC,
			},
			false,
			"",
			2,
		},
		{
			"without pagination",
			&tx.GetTxsEventRequest{
				Query: bankMsgSendEventAction,
			},
			false,
			"",
			2,
		},
		{
			"with pagination",
			&tx.GetTxsEventRequest{
				Query: bankMsgSendEventAction,
				Page:  1,
				Limit: 1,
			},
			false,
			"",
			1,
		},
		{
			"with multi events",
			&tx.GetTxsEventRequest{
				Query: fmt.Sprintf("%s AND message.module='bank'", bankMsgSendEventAction),
			},
			false,
			"",
			2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Query the tx via gRPC.
			grpcRes, err := qc.GetTxsEvent(context.Background(), tc.req)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(grpcRes.Txs), 1)
				require.Equal(t, "foobar", grpcRes.Txs[0].Body.Memo)
				require.Equal(t, tc.expLen, len(grpcRes.Txs))

				// Make sure fields are populated.
				require.NotEmpty(t, grpcRes.TxResponses[0].Timestamp)
				require.Empty(t, grpcRes.TxResponses[0].RawLog) // logs are empty if the transactions are successful
			}
		})
	}
}

func TestGetTxEvents_GRPCGateway(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	// qc := tx.NewServiceClient(sut.RPCClient(t))
	baseURL := systest.Sut.APIAddress()
	rsp := cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--note=foobar", "--fees=1stake")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)

	rsp = cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake")
	txResult, found = cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)

	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
		expLen    int
	}{
		{
			"empty params",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", baseURL),
			true,
			"query cannot be empty", 0,
		},
		{
			"without pagination",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s", baseURL, bankMsgSendEventAction),
			false,
			"", 2,
		},
		{
			"with pagination",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&page=%d&limit=%d", baseURL, bankMsgSendEventAction, 1, 1),
			false,
			"", 1,
		},
		{
			"valid request: order by asc",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&query=%s&order_by=ORDER_BY_ASC", baseURL, bankMsgSendEventAction, "message.module='bank'"),
			false,
			"", 2,
		},
		{
			"valid request: order by desc",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&query=%s&order_by=ORDER_BY_DESC", baseURL, bankMsgSendEventAction, "message.module='bank'"),
			false,
			"", 2,
		},
		{
			"invalid request: invalid order by",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&query=%s&order_by=invalid_order", baseURL, bankMsgSendEventAction, "message.module='bank'"),
			true,
			"is not a valid tx.OrderBy", 0,
		},
		{
			"expect pass with multiple-events",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&query=%s", baseURL, bankMsgSendEventAction, "message.module='bank'"),
			false,
			"", 2,
		},
		{
			"expect pass with escape event",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s", baseURL, "message.action%3D'/cosmos.bank.v1beta1.MsgSend'"),
			false,
			"", 2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := testutil.GetRequest(tc.url)
			require.NoError(t, err)
			if tc.expErr {
				require.Contains(t, string(res), tc.expErrMsg)
			} else {
				require.NoError(t, err)
				txs := gjson.Get(string(res), "txs").Array()
				require.Equal(t, len(txs), tc.expLen)
			}
		})
	}
}

func TestGetTx_GRPC(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	qc := tx.NewServiceClient(systest.Sut.RPCClient(t))

	rsp := cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake", "--note=foobar")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)
	txHash := gjson.Get(txResult, "txhash").String()

	testCases := []struct {
		name      string
		req       *tx.GetTxRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.GetTxRequest{}, true, "tx hash cannot be empty"},
		{"request with dummy hash", &tx.GetTxRequest{Hash: "deadbeef"}, true, "code = NotFound desc = tx not found: deadbeef"},
		{"good request", &tx.GetTxRequest{Hash: txHash}, false, ""},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Query the tx via gRPC.
			grpcRes, err := qc.GetTx(context.Background(), tc.req)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, "foobar", grpcRes.Tx.Body.Memo)
			}
		})
	}
}

func TestGetTx_GRPCGateway(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	baseURL := systest.Sut.APIAddress()

	rsp := cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake", "--note=foobar")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)
	txHash := gjson.Get(txResult, "txhash").String()

	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{
			"empty params",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/", baseURL),
			true, "tx hash cannot be empty",
		},
		{
			"dummy hash",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", baseURL, "deadbeef"),
			true, "tx not found",
		},
		{
			"good hash",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", baseURL, txHash),
			false, "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := testutil.GetRequest(tc.url)
			require.NoError(t, err)
			if tc.expErr {
				require.Contains(t, string(res), tc.expErrMsg)
			} else {
				timestamp := gjson.Get(string(res), "tx_response.timestamp").String()
				require.NotEmpty(t, timestamp)

				height := gjson.Get(string(res), "tx_response.height").Int()
				require.NotZero(t, height)

				rawLog := gjson.Get(string(res), "tx_response.raw_log").String()
				require.Empty(t, rawLog)
			}
		})
	}
}

func TestGetBlockWithTxs_GRPC(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	qc := tx.NewServiceClient(systest.Sut.RPCClient(t))

	rsp := cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake", "--note=foobar")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)
	height := gjson.Get(txResult, "height").Int()

	testCases := []struct {
		name      string
		req       *tx.GetBlockWithTxsRequest
		expErr    bool
		expErrMsg string
		expTxsLen int
	}{
		{"nil request", nil, true, "request cannot be nil", 0},
		{"empty request", &tx.GetBlockWithTxsRequest{}, true, "height must not be less than 1 or greater than the current height", 0},
		{"bad height", &tx.GetBlockWithTxsRequest{Height: 99999999}, true, "height must not be less than 1 or greater than the current height", 0},
		{"bad pagination", &tx.GetBlockWithTxsRequest{Height: height, Pagination: &query.PageRequest{Offset: 1000, Limit: 100}}, true, "out of range", 0},
		{"good request", &tx.GetBlockWithTxsRequest{Height: height}, false, "", 1},
		{"with pagination request", &tx.GetBlockWithTxsRequest{Height: height, Pagination: &query.PageRequest{Offset: 0, Limit: 1}}, false, "", 1},
		{"page all request", &tx.GetBlockWithTxsRequest{Height: height, Pagination: &query.PageRequest{Offset: 0, Limit: 100}}, false, "", 1},
		{"block with 0 tx", &tx.GetBlockWithTxsRequest{Height: height - 1, Pagination: &query.PageRequest{Offset: 0, Limit: 100}}, false, "", 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Query the tx via gRPC.
			grpcRes, err := qc.GetBlockWithTxs(context.Background(), tc.req)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
				if tc.expTxsLen > 0 {
					require.Equal(t, "foobar", grpcRes.Txs[0].Body.Memo)
				}
				require.Equal(t, grpcRes.Block.Header.Height, tc.req.Height)
				if tc.req.Pagination != nil {
					require.LessOrEqual(t, len(grpcRes.Txs), int(tc.req.Pagination.Limit))
				}
			}
		})
	}
}

func TestGetBlockWithTxs_GRPCGateway(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	baseUrl := systest.Sut.APIAddress()

	rsp := cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake", "--note=foobar")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)
	height := gjson.Get(txResult, "height").Int()

	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{
			"empty params",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/block/0", baseUrl),
			true, "height must not be less than 1 or greater than the current height",
		},
		{
			"bad height",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/block/%d", baseUrl, 9999999),
			true, "height must not be less than 1 or greater than the current height",
		},
		{
			"good request",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/block/%d", baseUrl, height),
			false, "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := testutil.GetRequest(tc.url)
			require.NoError(t, err)
			if tc.expErr {
				require.Contains(t, string(res), tc.expErrMsg)
			} else {
				memo := gjson.Get(string(res), "txs.0.body.memo").String()
				require.Equal(t, memo, "foobar")

				respHeight := gjson.Get(string(res), "block.header.height").Int()
				require.Equal(t, respHeight, height)
			}
		})
	}
}

func TestTxEncode_GRPC(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	systest.Sut.StartChain(t)

	qc := tx.NewServiceClient(systest.Sut.RPCClient(t))

	protoTx := &tx.Tx{
		Body: &tx.TxBody{
			Messages: []*codectypes.Any{},
		},
		AuthInfo:   &tx.AuthInfo{},
		Signatures: [][]byte{},
	}

	testCases := []struct {
		name      string
		req       *tx.TxEncodeRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxEncodeRequest{}, true, "invalid empty tx"},
		{"valid tx request", &tx.TxEncodeRequest{Tx: protoTx}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := qc.TxEncode(context.Background(), tc.req)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
				require.Empty(t, res)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTxEncode_GRPCGateway(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	systest.Sut.StartChain(t)

	baseUrl := systest.Sut.APIAddress()

	protoTx := &tx.Tx{
		Body: &tx.TxBody{
			Messages: []*codectypes.Any{},
		},
		AuthInfo:   &tx.AuthInfo{},
		Signatures: [][]byte{},
	}

	testCases := []struct {
		name      string
		req       *tx.TxEncodeRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxEncodeRequest{}, true, "invalid empty tx"},
		{"valid tx request", &tx.TxEncodeRequest{Tx: protoTx}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBz, err := json.Marshal(tc.req)
			require.NoError(t, err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/encode", baseUrl), "application/json", reqBz)
			require.NoError(t, err)
			if tc.expErr {
				require.Contains(t, string(res), tc.expErrMsg)
			}
		})
	}
}

func TestTxDecode_GRPC(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	qc := tx.NewServiceClient(systest.Sut.RPCClient(t))

	// create unsign tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + cli.ChainID(), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), "--from="+valAddr, "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home="+systest.Sut.NodeDir(0))
	signedTxFile := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "encode", signedTxFile.Name())
	txBz, err := base64.StdEncoding.DecodeString(res)
	require.NoError(t, err)
	invalidTxBytes := append(txBz, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxDecodeRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeRequest{TxBytes: invalidTxBytes}, true, "tx parse error"},
		{"valid request with tx bytes", &tx.TxDecodeRequest{TxBytes: txBz}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := qc.TxDecode(context.Background(), tc.req)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
				require.Empty(t, res)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, res.GetTx())
			}
		})
	}
}

func TestTxDecode_GRPCGateway(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	basrUrl := systest.Sut.APIAddress()

	// create unsign tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + cli.ChainID(), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), "--from="+valAddr, "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home="+systest.Sut.NodeDir(0))
	signedTxFile := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "encode", signedTxFile.Name())
	txBz, err := base64.StdEncoding.DecodeString(res)
	require.NoError(t, err)
	invalidTxBytes := append(txBz, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxDecodeRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeRequest{TxBytes: invalidTxBytes}, true, "tx parse error"},
		{"valid request with tx_bytes", &tx.TxDecodeRequest{TxBytes: txBz}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBz, err := json.Marshal(tc.req)
			require.NoError(t, err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/decode", basrUrl), "application/json", reqBz)
			require.NoError(t, err)
			if tc.expErr {
				require.Contains(t, string(res), tc.expErrMsg)
			} else {
				signatures := gjson.Get(string(res), "tx.signatures").Array()
				require.Equal(t, len(signatures), 1)
			}
		})
	}
}

func TestTxEncodeAmino_GRPC(t *testing.T) {
	systest.Sut.ResetChain(t)
	systest.Sut.StartChain(t)

	legacyAmino := codec.NewLegacyAmino()
	std.RegisterLegacyAminoCodec(legacyAmino)
	legacytx.RegisterLegacyAminoCodec(legacyAmino)
	legacy.RegisterAminoMsg(legacyAmino, &banktypes.MsgSend{}, "cosmos-sdk/MsgSend")

	qc := tx.NewServiceClient(systest.Sut.RPCClient(t))
	txJSONBytes, stdTx := readTestAminoTxJSON(t, legacyAmino)

	testCases := []struct {
		name      string
		req       *tx.TxEncodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxEncodeAminoRequest{}, true, "invalid empty tx json"},
		{"invalid request", &tx.TxEncodeAminoRequest{AminoJson: "invalid tx json"}, true, "invalid request"},
		{"valid request with amino-json", &tx.TxEncodeAminoRequest{AminoJson: string(txJSONBytes)}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := qc.TxEncodeAmino(context.Background(), tc.req)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
				require.Empty(t, res)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, res.GetAminoBinary())

				var decodedTx legacytx.StdTx
				err = legacyAmino.Unmarshal(res.AminoBinary, &decodedTx)
				require.NoError(t, err)
				require.Equal(t, decodedTx.GetMsgs(), stdTx.GetMsgs())
			}
		})
	}
}

func TestTxEncodeAmino_GRPCGateway(t *testing.T) {
	systest.Sut.ResetChain(t)
	systest.Sut.StartChain(t)

	legacyAmino := codec.NewLegacyAmino()
	std.RegisterLegacyAminoCodec(legacyAmino)
	legacytx.RegisterLegacyAminoCodec(legacyAmino)
	legacy.RegisterAminoMsg(legacyAmino, &banktypes.MsgSend{}, "cosmos-sdk/MsgSend")

	baseUrl := systest.Sut.APIAddress()
	txJSONBytes, stdTx := readTestAminoTxJSON(t, legacyAmino)

	testCases := []struct {
		name      string
		req       *tx.TxEncodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxEncodeAminoRequest{}, true, "invalid empty tx json"},
		{"invalid request", &tx.TxEncodeAminoRequest{AminoJson: "invalid tx json"}, true, "cannot parse disfix JSON wrapper"},
		{"valid request with amino-json", &tx.TxEncodeAminoRequest{AminoJson: string(txJSONBytes)}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBz, err := json.Marshal(tc.req)
			require.NoError(t, err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/encode/amino", baseUrl), "application/json", reqBz)
			require.NoError(t, err)
			if tc.expErr {
				require.Contains(t, string(res), tc.expErrMsg)
			} else {
				var result tx.TxEncodeAminoResponse
				err := json.Unmarshal(res, &result)
				require.NoError(t, err)

				var decodedTx legacytx.StdTx
				err = legacyAmino.Unmarshal(result.AminoBinary, &decodedTx)
				require.NoError(t, err)
				require.Equal(t, decodedTx.GetMsgs(), stdTx.GetMsgs())
			}
		})
	}
}

func TestTxDecodeAmino_GRPC(t *testing.T) {
	systest.Sut.ResetChain(t)
	systest.Sut.StartChain(t)

	legacyAmino := codec.NewLegacyAmino()
	std.RegisterLegacyAminoCodec(legacyAmino)
	legacytx.RegisterLegacyAminoCodec(legacyAmino)
	legacy.RegisterAminoMsg(legacyAmino, &banktypes.MsgSend{}, "cosmos-sdk/MsgSend")

	qc := tx.NewServiceClient(systest.Sut.RPCClient(t))
	encodedTx, stdTx := readTestAminoTxBinary(t, legacyAmino)

	invalidTxBytes := append(encodedTx, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxDecodeAminoRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: invalidTxBytes}, true, "unmarshal to legacytx.StdTx failed"},
		{"valid request with tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: encodedTx}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := qc.TxDecodeAmino(context.Background(), tc.req)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
				require.Empty(t, res)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, res.GetAminoJson())

				var decodedTx legacytx.StdTx
				err = legacyAmino.UnmarshalJSON([]byte(res.GetAminoJson()), &decodedTx)
				require.NoError(t, err)
				require.Equal(t, stdTx.GetMsgs(), decodedTx.GetMsgs())
			}
		})
	}
}

func TestTxDecodeAmino_GRPCGateway(t *testing.T) {
	systest.Sut.ResetChain(t)
	systest.Sut.StartChain(t)

	legacyAmino := codec.NewLegacyAmino()
	std.RegisterLegacyAminoCodec(legacyAmino)
	legacytx.RegisterLegacyAminoCodec(legacyAmino)
	legacy.RegisterAminoMsg(legacyAmino, &banktypes.MsgSend{}, "cosmos-sdk/MsgSend")

	baseUrl := systest.Sut.APIAddress()
	encodedTx, stdTx := readTestAminoTxBinary(t, legacyAmino)

	invalidTxBytes := append(encodedTx, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxDecodeAminoRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: invalidTxBytes}, true, "unmarshal to legacytx.StdTx failed"},
		{"valid request with tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: encodedTx}, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBz, err := json.Marshal(tc.req)
			require.NoError(t, err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/decode/amino", baseUrl), "application/json", reqBz)
			require.NoError(t, err)
			if tc.expErr {
				require.Contains(t, string(res), tc.expErrMsg)
			} else {
				var result tx.TxDecodeAminoResponse
				err := json.Unmarshal(res, &result)
				require.NoError(t, err)

				var decodedTx legacytx.StdTx
				err = legacyAmino.UnmarshalJSON([]byte(result.AminoJson), &decodedTx)
				require.NoError(t, err)
				require.Equal(t, stdTx.GetMsgs(), decodedTx.GetMsgs())
			}
		})
	}
}

func TestSimMultiSigTx(t *testing.T) {
	t.Skip() // waiting for @hieuvubk fix

	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	_ = cli.AddKey("account1")
	_ = cli.AddKey("account2")

	systest.Sut.StartChain(t)

	multiSigName := "multisig"
	cli.RunCommandWithArgs("keys", "add", multiSigName, "--multisig=account1,account2", "--multisig-threshold=2", "--keyring-backend=test", "--home=./testnet")
	multiSigAddr := cli.GetKeyAddr(multiSigName)

	// Send from validator to multisig addr
	rsp := cli.Run("tx", "bank", "send", valAddr, multiSigAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)

	multiSigBalance := cli.QueryBalance(multiSigAddr, denom)
	require.Equal(t, multiSigBalance, transferAmount)

	// Send from multisig to validator
	// create unsign tx
	var newTransferAmount int64 = 100
	bankSendCmdArgs := []string{"tx", "bank", "send", multiSigAddr, valAddr, fmt.Sprintf("%d%s", newTransferAmount, denom), "--chain-id=" + cli.ChainID(), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), fmt.Sprintf("--from=%s", "account1"), fmt.Sprintf("--multisig=%s", multiSigAddr), "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home=./testnet")
	account1Signed := systest.StoreTempFile(t, []byte(res))
	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), fmt.Sprintf("--from=%s", "account2"), fmt.Sprintf("--multisig=%s", multiSigAddr), "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home=./testnet")
	account2Signed := systest.StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "multisign-batch", txFile.Name(), multiSigName, account1Signed.Name(), account2Signed.Name(), "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home=./testnet")
	txSignedFile := systest.StoreTempFile(t, []byte(res))

	res = cli.Run("tx", "broadcast", txSignedFile.Name())
	systest.RequireTxSuccess(t, res)

	multiSigBalance = cli.QueryBalance(multiSigAddr, denom)
	require.Equal(t, multiSigBalance, transferAmount-newTransferAmount-10)
}

func readTestAminoTxJSON(t *testing.T, aminoCodec *codec.LegacyAmino) ([]byte, *legacytx.StdTx) {
	txJSONBytes, err := os.ReadFile("testdata/tx_amino1.json")
	require.NoError(t, err)
	var stdTx legacytx.StdTx
	err = aminoCodec.UnmarshalJSON(txJSONBytes, &stdTx)
	require.NoError(t, err)
	return txJSONBytes, &stdTx
}

func readTestAminoTxBinary(t *testing.T, aminoCodec *codec.LegacyAmino) ([]byte, *legacytx.StdTx) {
	txJSONBytes, err := os.ReadFile("testdata/tx_amino1.bin")
	require.NoError(t, err)
	var stdTx legacytx.StdTx
	err = aminoCodec.Unmarshal(txJSONBytes, &stdTx)
	require.NoError(t, err)
	return txJSONBytes, &stdTx
}

func TestSimulateTx_GasImprovements(t *testing.T) {
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	// use node 1 as granter
	granter := cli.GetKeyAddr("node1")

	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	systest.Sut.StartChain(t)

	baseURL := systest.Sut.APIAddress()

	// node1 grant to node0
	grantCmdArgs := []string{"tx", "feegrant", "grant", granter, valAddr, "--chain-id=" + cli.ChainID()}
	rsp := cli.Run(grantCmdArgs...)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)

	sendSimulateCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + cli.ChainID(), "--sign-mode=direct", "--generate-only"}
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + cli.ChainID()}
	testCases := []struct {
		name         string
		simulateArgs []string
		txArgs []string
	}{
		{
			"simulate without fees",
			sendSimulateCmdArgs,
			bankSendCmdArgs,
		},
		{
			"simulate with fees",
			append(sendSimulateCmdArgs, "--fees=1stake"),
			bankSendCmdArgs,
		},
		{
			"simulate without fee-granter",
			sendSimulateCmdArgs,
			append(bankSendCmdArgs, fmt.Sprintf("--fee-granter=%s", granter)),
		},
		{
			"simulate with fee-granter",
			append(sendSimulateCmdArgs, fmt.Sprintf("--fee-granter=%s", granter)),
			append(bankSendCmdArgs, fmt.Sprintf("--fee-granter=%s", granter)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txlen := 1
			gasSimulated := make([]int64, txlen)
			gasAdjustment := make([]float64, txlen)

			for i := 0; i < txlen; i++ {
				// create unsign tx
				res := cli.RunCommandWithArgs(tc.simulateArgs...)
				txFile := systest.StoreTempFile(t, []byte(res))
		
				res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), "--from="+valAddr, "--chain-id="+cli.ChainID(), "--keyring-backend=test", "--home="+systest.Sut.NodeDir(0))
				signedTxFile := systest.StoreTempFile(t, []byte(res))
		
				res = cli.RunCommandWithArgs("tx", "encode", signedTxFile.Name())
				txBz, err := base64.StdEncoding.DecodeString(res)
				require.NoError(t, err)
		
				reqBz, err := json.Marshal(&tx.SimulateRequest{TxBytes: txBz})
				require.NoError(t, err)
		
				resBz, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/simulate", baseURL), "application/json", reqBz)
				require.NoError(t, err)
				gasUsed := gjson.Get(string(resBz), "gas_info.gas_used").Int()
				require.True(t, gasUsed > 0)
				gasSimulated[i] = gasUsed
			}

			for i := 0; i < txlen; i++ {
				// Submit tx with gas return from simulate
				gasAdjustment[i] = 0.99
				// test valid transaction
				shouldRerun := true
				for shouldRerun {
					gasAdjustment[i] = gasAdjustment[i] + 0.01
					gasUse := int64(gasAdjustment[i] * float64(gasSimulated[i]))
					rsp := cli.Run(append(tc.txArgs, fmt.Sprintf("--gas=%d", gasUse))...)
					// rerun if tx err and increase gas-adjustment by 1%
					shouldRerun = strings.Contains(gjson.Get(rsp, "raw_log").String(), "out of gas") || gjson.Get(rsp, "gas_used").Int() == int64(0) || gjson.Get(rsp, "code").Int() != 0
					fmt.Println("gasAdjustment", i, gasAdjustment[i])
				}
			}
		
			// Calculate average adjustments
			total := 0.0
			for i := 0; i < txlen; i++ {
				total = total + gasAdjustment[i]
			}
			average := total / float64(txlen)
			result := fmt.Sprintf("Test case %s average gas adjustment is: %f", tc.name, average)
			fmt.Println(result)
		})
	}
}
