//go:build system_test

package systemtests

import (
	"context"
	"encoding/base64"

	"fmt"

	"testing"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var bankMsgSendEventAction = "message.action='/cosmos.bank.v1beta1.MsgSend'"

func TestQueryBySig(t *testing.T) {
	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	var transferAmount int64 = 1000

	sut.StartChain(t)

	qc := tx.NewServiceClient(sut.RPCClient(t))

	// create unsign tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), fmt.Sprintf("--from=%s", valAddr), fmt.Sprintf("--chain-id=%s", sut.chainID), "--keyring-backend=test", "--home=./testnet/node0/simd")
	sig := gjson.Get(res, "signatures.0").String()
	signedTxFile := StoreTempFile(t, []byte(res))

	res = cli.Run("tx", "broadcast", signedTxFile.Name())
	RequireTxSuccess(t, res)

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
	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	var transferAmount int64 = 1000

	sut.StartChain(t)

	qc := tx.NewServiceClient(sut.RPCClient(t))

	// create unsign tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), fmt.Sprintf("--from=%s", valAddr), fmt.Sprintf("--chain-id=%s", sut.chainID), "--keyring-backend=test", "--home=./testnet/node0/simd")
	signedTxFile := StoreTempFile(t, []byte(res))

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
				// The 12 events are:
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

func TestGetTxEvents_GRPC(t *testing.T) {
	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	var transferAmount int64 = 1000

	sut.StartChain(t)

	qc := tx.NewServiceClient(sut.RPCClient(t))
	rsp := cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--note=foobar", "--fees=1stake")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

	rsp = cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake")
	txResult, found = cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

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
				// ref: https://github.com/cosmos/cosmos-sdk/issues/8680
				// ref: https://github.com/cosmos/cosmos-sdk/issues/8681
				require.NotEmpty(t, grpcRes.TxResponses[0].Timestamp)
				require.Empty(t, grpcRes.TxResponses[0].RawLog) // logs are empty if the transactions are successful
			}
		})
	}
	


}

func TestGetTx_GRPC(t *testing.T) {
	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	var transferAmount int64 = 1000

	sut.StartChain(t)

	qc := tx.NewServiceClient(sut.RPCClient(t))

	rsp := cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake", "--note=foobar")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)
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

func TestGetBlockWithTxs_GRPC(t *testing.T) {
	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	var transferAmount int64 = 1000

	sut.StartChain(t)

	qc := tx.NewServiceClient(sut.RPCClient(t))

	rsp := cli.Run("tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake", "--note=foobar")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)
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
		t.Run(tc.name, func(t * testing.T) {
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

func TestTxEncode_GRPC(t *testing.T) {
	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	sut.StartChain(t)

	qc := tx.NewServiceClient(sut.RPCClient(t))

	protoTx := &tx.Tx{
		Body: &tx.TxBody{
			Messages: []*codectypes.Any{},
		},
		AuthInfo: &tx.AuthInfo{},
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

func TestTxDecode_GRPC(t *testing.T) {
	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	var transferAmount int64 = 1000

	sut.StartChain(t)

	qc := tx.NewServiceClient(sut.RPCClient(t))

	// create unsign tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), fmt.Sprintf("--from=%s", valAddr), fmt.Sprintf("--chain-id=%s", sut.chainID), "--keyring-backend=test", "--home=./testnet/node0/simd")
	signedTxFile := StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "encode", signedTxFile.Name())
	txBz, err := base64.StdEncoding.DecodeString(res)
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

func TestSimMultiSigTx(t *testing.T) {
	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	_ = cli.AddKey("account1")
	_ = cli.AddKey("account2")
	denom := "stake"
	var transferAmount int64 = 1000

	sut.StartChain(t)

	multiSigName := "multisig"
	res := cli.RunCommandWithArgs("keys", "add", multiSigName, "--multisig=account1,account2", "--multisig-threshold=2", "--keyring-backend=test", "--home=./testnet")
	multiSigAddr := cli.GetKeyAddr(multiSigName)

	// Send from validator to multisig addr
	rsp := cli.Run("tx", "bank", "send", valAddr, multiSigAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=1stake")
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

	multiSigBalance := cli.QueryBalance(multiSigAddr, denom)
	require.Equal(t, multiSigBalance, transferAmount)

	// Send from multisig to validator
	// create unsign tx
	var newTransferAmount int64 = 100
	bankSendCmdArgs := []string{"tx", "bank", "send", multiSigAddr, valAddr, fmt.Sprintf("%d%s", newTransferAmount, denom), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res = cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), fmt.Sprintf("--from=%s", "account1"), fmt.Sprintf("--multisig=%s", multiSigAddr), fmt.Sprintf("--chain-id=%s", sut.chainID), "--keyring-backend=test", "--home=./testnet")
	account1Signed := StoreTempFile(t, []byte(res))
	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), fmt.Sprintf("--from=%s", "account2"), fmt.Sprintf("--multisig=%s", multiSigAddr), fmt.Sprintf("--chain-id=%s", sut.chainID), "--keyring-backend=test", "--home=./testnet")
	account2Signed := StoreTempFile(t, []byte(res))

	res = cli.RunCommandWithArgs("tx", "multisign-batch", txFile.Name(), multiSigName, account1Signed.Name(), account2Signed.Name(), fmt.Sprintf("--chain-id=%s", sut.chainID), "--keyring-backend=test", "--home=./testnet")
	txSignedFile := StoreTempFile(t, []byte(res))

	res = cli.Run("tx", "broadcast", txSignedFile.Name())
	RequireTxSuccess(t, res)
	
	multiSigBalance = cli.QueryBalance(multiSigAddr, denom)
	require.Equal(t, multiSigBalance, transferAmount - newTransferAmount - 10)

	
}