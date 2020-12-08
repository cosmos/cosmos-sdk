package cosmos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/go-amino"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type DataClient struct {
	tm        rpcclient.Client
	lcd       string
	cdc       *amino.Codec
	txDecoder sdk.TxDecoder
}

func (d DataClient) SupportedOperations() []string {
	return supportedOperations
}

func (d DataClient) NodeVersion() string {
	return "0.37.12"
}

func NewDataClient(tmEndpoint string, lcdEndpoint string, cdc *amino.Codec) (DataClient, error) {
	tmClient := rpcclient.NewHTTP(tmEndpoint, "/websocket")
	// test it works
	_, err := tmClient.Health()
	if err != nil {
		return DataClient{}, err
	}
	dc := DataClient{
		tm:        tmClient,
		lcd:       lcdEndpoint,
		cdc:       cdc,
		txDecoder: auth.DefaultTxDecoder(cdc),
	}
	return dc, nil
}

func (d DataClient) Balances(ctx context.Context, address string, height *int64) (amounts []*types.Amount, err error) {
	balance, err := d.balance(ctx, address, height)
	if err != nil {
		return
	}

	amounts = make([]*types.Amount, len(balance))
	for i, coin := range balance {
		amounts[i] = &types.Amount{
			Value: coin.Amount.String(),
			Currency: &types.Currency{
				Symbol: coin.Denom,
			},
		}
	}
	return
}

func (d DataClient) do(ctx context.Context, path string, height *int64, req interface{}, resp interface{}) error {
	u := fmt.Sprintf("%s/%s", d.lcd, path)
	if height != nil {
		u += "?height=" + strconv.FormatInt(*height, 10)
	}

	// in case a body is provided then marshal it and use it as replace io.Reader
	// otherwise the body will be nil and ignored by the request doer
	var body io.Reader = nil
	if req != nil {
		reqBody, err := d.cdc.MarshalJSON(req)
		if err != nil {
			return rosetta.WrapError(rosetta.ErrCodec, err.Error())
		}
		body = bytes.NewBuffer(reqBody)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, body)
	if err != nil {
		return rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	defer httpResp.Body.Close()

	rawBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}

	var x map[string]json.RawMessage
	err = d.cdc.UnmarshalJSON(rawBody, &x)
	if err != nil {
		return rosetta.WrapError(rosetta.ErrCodec, err.Error())
	}

	queryResult, ok := x["result"]
	if !ok {
		return rosetta.WrapError(rosetta.ErrCodec, "result missing from query response")
	}
	err = d.cdc.UnmarshalJSON(queryResult, resp)
	return nil
}

func (d DataClient) balance(ctx context.Context, address string, height *int64) (coins sdk.Coins, err error) {
	const path = "bank/balances"
	sdkAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrBadArgument, err.Error())
	}
	queryPath := fmt.Sprintf("%s/%s", path, sdkAddr.String())
	err = d.do(ctx, queryPath, height, nil, &coins)
	return
}

func (d DataClient) supply(ctx context.Context, height *int64) (coins sdk.Coins, err error) {
	const path = "supply/total_supply"
	supplyReq := struct {
		Page, Limit int
	}{}
	err = d.do(ctx, path, height, supplyReq, &coins)
	return
}

func (d DataClient) BlockByHeight(_ context.Context, height *int64) (block rosetta.BlockResponse, err error) {
	tmBlock, err := d.tm.Block(height)
	if err != nil {
		return block, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	block = rosetta.BlockResponse{
		Block: &types.BlockIdentifier{
			Index: tmBlock.BlockMeta.Header.Height,
			Hash:  fmt.Sprintf("%X", tmBlock.Block.Hash()),
		},
		ParentBlock: &types.BlockIdentifier{
			Index: tmBlock.BlockMeta.Header.Height - 1,
			Hash:  fmt.Sprintf("%X", tmBlock.BlockMeta.Header.LastBlockID.Hash),
		},
		MillisecondTimestamp: timeToMilliseconds(tmBlock.Block.Time),
		TxCount:              tmBlock.Block.NumTxs,
	}
	return block, nil
}

func (d DataClient) BlockByHash(_ context.Context, _ string) (block rosetta.BlockResponse, err error) {
	return block, rosetta.WrapError(rosetta.ErrNotImplemented, "unable to get block by hash")
}

func (d DataClient) BlockTransactionsByHeight(ctx context.Context, height *int64) (block rosetta.BlockTransactionsResponse, err error) {
	tmBlock, err := d.BlockByHeight(ctx, height)
	if err != nil {
		return block, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}

	// set up block
	block.BlockResponse = tmBlock
	// if the txs in the block are 0 then return
	if block.TxCount == 0 {
		return block, nil
	}
	// otherwise fetch transactions and add them to block
	tmTxs, err := d.tm.TxSearch(fmt.Sprintf("tx.height=%d", tmBlock.Block.Index), true, 0, 0)
	if err != nil {
		return block, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}

	block.Transactions = make([]*types.Transaction, tmBlock.TxCount)
	for i, tmTx := range tmTxs.Txs {
		decodedTx, err := d.txDecoder(tmTx.Tx)
		if err != nil {
			return block, err
		}
		block.Transactions[i] = &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{Hash: fmt.Sprintf("%X", tmTx.Hash)},
			Operations:            sdkTxToOperations(decodedTx, true, tmTx.TxResult.Code != 0),
		}

	}
	return block, nil
}

func (d DataClient) BlockTransactionsByHash(_ context.Context, _ string) (block rosetta.BlockTransactionsResponse, err error) {
	return block, rosetta.WrapError(rosetta.ErrNotImplemented, "unable to get block transactions given a block hash")
}

func (d DataClient) GetTransaction(_ context.Context, hash string) (tx *types.Transaction, err error) {
	tmTx, err := d.tm.Tx([]byte(hash), false)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	var cosmosTx sdk.TxResponse
	err = d.cdc.UnmarshalBinaryBare(tmTx.TxResult.Data, &cosmosTx)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrCodec, err.Error())
	}
	tx = &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: hash},
		Operations:            sdkTxToOperations(cosmosTx.Tx, true, cosmosTx.Code != 0),
	}
	return tx, nil
}

func (d DataClient) GetMempoolTransactions(_ context.Context) (txs []*types.TransactionIdentifier, err error) {
	tmTxs, err := d.tm.UnconfirmedTxs(0)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	txs = make([]*types.TransactionIdentifier, len(tmTxs.Txs))
	for i, tmTx := range tmTxs.Txs {
		txs[i] = &types.TransactionIdentifier{Hash: fmt.Sprintf("%X", tmTx.Hash())}
	}
	return
}

func (d DataClient) GetMempoolTransaction(_ context.Context, _ string) (tx *types.Transaction, err error) {
	return nil, rosetta.ErrNotImplemented
}

func (d DataClient) Peers(_ context.Context) (peers []*types.Peer, err error) {
	netInfo, err := d.tm.NetInfo()
	if err != nil {
		return peers, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	peers = make([]*types.Peer, len(netInfo.Peers))
	for i, tmPeer := range netInfo.Peers {
		peers[i] = &types.Peer{
			PeerID: (string)(tmPeer.NodeInfo.ID()),
		}
	}
	return
}

func (d DataClient) Status(_ context.Context) (status *types.SyncStatus, err error) {
	tmStatus, err := d.tm.Status()
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	status = &types.SyncStatus{
		CurrentIndex: tmStatus.SyncInfo.LatestBlockHeight,
		TargetIndex:  nil,
		Stage:        nil,
	}
	return
}
