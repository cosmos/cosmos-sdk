package rosetta

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/tendermint/btcd/btcec"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/coinbase/rosetta-sdk-go/types"
	"google.golang.org/grpc/metadata"

	"github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"

	crgerrs "github.com/tendermint/cosmos-rosetta-gateway/errors"
	crgtypes "github.com/tendermint/cosmos-rosetta-gateway/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// interface assertion
var _ crgtypes.Client = (*Client)(nil)

const tmWebsocketPath = "/websocket"
const defaultNodeTimeout = 15 * time.Second

// Client implements a single network client to interact with cosmos based chains
type Client struct {
	config *Config

	auth auth.QueryClient
	bank bank.QueryClient

	ir codectypes.InterfaceRegistry

	clientCtx client.Context
}

func (c *Client) AccountIdentifierFromPublicKey(pubKey *types.PublicKey) (*types.AccountIdentifier, error) {
	if pubKey.CurveType != "secp256k1" {
		return nil, crgerrs.WrapError(crgerrs.ErrUnsupportedCurve, "only secp256k1 supported")
	}

	cmp, err := btcec.ParsePubKey(pubKey.Bytes, btcec.S256())
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, err.Error())
	}

	compressedPublicKey := make([]byte, secp256k1.PubKeySize)
	copy(compressedPublicKey, cmp.SerializeCompressed())

	pk := secp256k1.PubKey{Key: compressedPublicKey}

	return &types.AccountIdentifier{
		Address: sdk.AccAddress(pk.Address()).String(),
	}, nil
}

// NewClient instantiates a new online servicer
func NewClient(cfg *Config) (*Client, error) {
	return &Client{
		config: cfg,
		ir:     cfg.InterfaceRegistry,
	}, nil
}

func (c *Client) accountInfo(ctx context.Context, addr string, height *int64) (auth.AccountI, error) {
	if height != nil {
		strHeight := strconv.FormatInt(*height, 10)
		ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	}

	accountInfo, err := c.auth.Account(ctx, &auth.QueryAccountRequest{
		Address: addr,
	})
	if err != nil {
		return nil, crgerrs.FromGRPCToRosettaError(err)
	}

	var account auth.AccountI
	err = c.ir.UnpackAny(accountInfo.Account, &account)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, err.Error())
	}

	return account, nil
}

func (c *Client) Balances(ctx context.Context, addr string, height *int64) ([]*types.Amount, error) {
	if height != nil {
		strHeight := strconv.FormatInt(*height, 10)
		ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	}

	balance, err := c.bank.AllBalances(ctx, &bank.QueryAllBalancesRequest{
		Address: addr,
	})
	if err != nil {
		return nil, crgerrs.FromGRPCToRosettaError(err)
	}

	availableCoins, err := c.coins(ctx)
	if err != nil {
		return nil, err
	}

	return sdkCoinsToRosettaAmounts(balance.Balances, availableCoins), nil
}

func (c *Client) BlockByHash(ctx context.Context, hash string) (crgtypes.BlockResponse, error) {
	bHash, err := hex.DecodeString(hash)
	if err != nil {
		return crgtypes.BlockResponse{}, fmt.Errorf("invalid block hash: %s", err)
	}

	block, err := c.clientCtx.Client.BlockByHash(ctx, bHash)
	if err != nil {
		return crgtypes.BlockResponse{}, err
	}

	return buildBlockResponse(block), nil
}

func (c *Client) BlockByHeight(ctx context.Context, height *int64) (crgtypes.BlockResponse, error) {
	block, err := c.clientCtx.Client.Block(ctx, height)
	if err != nil {
		return crgtypes.BlockResponse{}, err
	}

	return buildBlockResponse(block), nil
}

func buildBlockResponse(block *tmtypes.ResultBlock) crgtypes.BlockResponse {
	return crgtypes.BlockResponse{
		Block:                TMBlockToRosettaBlockIdentifier(block),
		ParentBlock:          TMBlockToRosettaParentBlockIdentifier(block),
		MillisecondTimestamp: timeToMilliseconds(block.Block.Time),
		TxCount:              int64(len(block.Block.Txs)),
	}
}

func (c *Client) BlockTransactionsByHash(ctx context.Context, hash string) (crgtypes.BlockTransactionsResponse, error) {
	blockResp, err := c.BlockByHash(ctx, hash)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}

	txs, err := c.listTransactionsInBlock(ctx, blockResp.Block.Index)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}

	return crgtypes.BlockTransactionsResponse{
		BlockResponse: blockResp,
		Transactions:  sdkTxsWithHashToRosettaTxs(txs),
	}, nil
}

func (c *Client) BlockTransactionsByHeight(ctx context.Context, height *int64) (crgtypes.BlockTransactionsResponse, error) {
	blockResp, err := c.BlockByHeight(ctx, height)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}

	txs, err := c.listTransactionsInBlock(ctx, blockResp.Block.Index)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}

	return crgtypes.BlockTransactionsResponse{
		BlockResponse: blockResp,
		Transactions:  sdkTxsWithHashToRosettaTxs(txs),
	}, nil
}

// Coins fetches the existing coins in the application
func (c *Client) coins(ctx context.Context) (sdk.Coins, error) {
	supply, err := c.bank.TotalSupply(ctx, &bank.QueryTotalSupplyRequest{})
	if err != nil {
		return nil, crgerrs.FromGRPCToRosettaError(err)
	}
	return supply.Supply, nil
}

// listTransactionsInBlock returns the list of the transactions in a block given its height
func (c *Client) listTransactionsInBlock(ctx context.Context, height int64) ([]*sdkTxWithHash, error) {
	txQuery := fmt.Sprintf(`tx.height=%d`, height)
	txList, err := c.clientCtx.Client.TxSearch(ctx, txQuery, true, nil, nil, "")
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}

	sdkTxs, err := tmResultTxsToSdkTxsWithHash(c.clientCtx.TxConfig.TxDecoder(), txList.Txs)
	if err != nil {
		return nil, err
	}
	return sdkTxs, nil
}

func (c *Client) TxOperationsAndSignersAccountIdentifiers(signed bool, txBytes []byte) (ops []*types.Operation, signers []*types.AccountIdentifier, err error) {
	txConfig := c.getTxConfig()
	rawTx, err := txConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, nil, err
	}
	txBldr, _ := txConfig.WrapTxBuilder(rawTx)

	var accountIdentifierSigners []*types.AccountIdentifier
	if signed {
		addrs := txBldr.GetTx().GetSigners()
		for _, addr := range addrs {
			signer := &types.AccountIdentifier{
				Address: addr.String(),
			}
			accountIdentifierSigners = append(accountIdentifierSigners, signer)
		}
	}

	return sdkTxToOperations(txBldr.GetTx(), false, false), accountIdentifierSigners, nil
}

// GetTx returns a transaction given its hash
func (c *Client) GetTx(_ context.Context, hash string) (*types.Transaction, error) {
	txResp, err := authclient.QueryTx(c.clientCtx, hash)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	var sdkTx sdk.Tx
	err = c.ir.UnpackAny(txResp.Tx, &sdkTx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, err.Error())
	}
	return sdkTxWithHashToOperations(&sdkTxWithHash{
		HexHash: txResp.TxHash,
		Code:    txResp.Code,
		Log:     txResp.RawLog,
		Tx:      sdkTx,
	}), nil
}

// GetUnconfirmedTx gets an unconfirmed transaction given its hash
// NOTE(fdymylja): not implemented yet
func (c *Client) GetUnconfirmedTx(_ context.Context, _ string) (*types.Transaction, error) {
	return nil, crgerrs.WrapError(crgerrs.ErrNotImplemented, "get unconfirmed transaction method is not supported")
}

// Mempool returns the unconfirmed transactions in the mempool
func (c *Client) Mempool(ctx context.Context) ([]*types.TransactionIdentifier, error) {
	txs, err := c.clientCtx.Client.UnconfirmedTxs(ctx, nil)
	if err != nil {
		return nil, err
	}

	return TMTxsToRosettaTxsIdentifiers(txs.Txs), nil
}

// Peers gets the number of peers
func (c *Client) Peers(ctx context.Context) ([]*types.Peer, error) {
	netInfo, err := c.clientCtx.Client.NetInfo(ctx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	return TmPeersToRosettaPeers(netInfo.Peers), nil
}

func (c *Client) Status(ctx context.Context) (*types.SyncStatus, error) {
	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	return TMStatusToRosettaSyncStatus(status), err
}

func (c *Client) getTxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}

func (c *Client) PostTx(txBytes []byte) (*types.TransactionIdentifier, map[string]interface{}, error) {
	res, err := c.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, nil, err
	}

	return &types.TransactionIdentifier{
			Hash: res.TxHash,
		},
		map[string]interface{}{
			Log: res.RawLog,
		}, nil
}

func (c *Client) ConstructionMetadataFromOptions(ctx context.Context, options map[string]interface{}) (meta map[string]interface{}, err error) {
	if len(options) == 0 {
		return nil, crgerrs.ErrBadArgument
	}

	addr, ok := options[OptionAddress]
	if !ok {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidAddress, "no address provided")
	}

	addrString := addr.(string)

	accountInfo, err := c.accountInfo(ctx, addrString, nil)
	if err != nil {
		return nil, err
	}

	gas, ok := options[OptionGas]
	if !ok {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidAddress, "gas not set")
	}

	memo, ok := options[OptionMemo]
	if !ok {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidMemo, "memo not set")
	}

	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		OptionAccountNumber: accountInfo.GetAccountNumber(),
		OptionSequence:      accountInfo.GetSequence(),
		OptionChainID:       status.NodeInfo.Network,
		OptionGas:           gas,
		OptionMemo:          memo,
	}, nil
}

func (c *Client) Ready() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultNodeTimeout)
	defer cancel()
	_, err := c.clientCtx.Client.Health(ctx)
	if err != nil {
		return err
	}
	_, err = c.bank.TotalSupply(ctx, &bank.QueryTotalSupplyRequest{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Bootstrap() error {
	grpcConn, err := grpc.Dial(c.config.GRPCEndpoint, grpc.WithInsecure())
	if err != nil {
		return err
	}

	tmRPC, err := http.New(c.config.TendermintRPC, tmWebsocketPath)
	if err != nil {
		return err
	}

	authClient := auth.NewQueryClient(grpcConn)
	bankClient := bank.NewQueryClient(grpcConn)

	// NodeURI and Client are set from here otherwise
	// WitNodeURI will require to create a new client
	// it's done here because WithNodeURI panics if
	// connection to tendermint node fails
	clientCtx := client.Context{
		Client:  tmRPC,
		NodeURI: c.config.TendermintRPC,
	}
	clientCtx = clientCtx.
		WithJSONMarshaler(c.config.Codec).
		WithInterfaceRegistry(c.config.InterfaceRegistry).
		WithTxConfig(authtx.NewTxConfig(c.config.Codec, authtx.DefaultSignModes)).
		WithAccountRetriever(auth.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock)

	c.auth = authClient
	c.bank = bankClient
	c.clientCtx = clientCtx
	c.ir = c.config.InterfaceRegistry

	return nil
}
