package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"google.golang.org/grpc/metadata"

	"github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// interface assertion
var _ rosetta.NodeClient = (*Client)(nil)

const tmWebsocketPath = "/websocket"

// options defines optional settings for SingleClient
type options struct {
	interfaceRegistry codectypes.InterfaceRegistry
	cdc               *codec.ProtoCodec
}

// newDefaultOptions builds the default options
func newDefaultOptions() options {
	cdc, ir := MakeCodec()
	return options{
		interfaceRegistry: ir,
		cdc:               cdc,
	}
}

// OptionFunc defines a function that edits single client options
type OptionFunc func(o *options)

// WithCodec allows to build a client with a predefined interface registry and codec
func WithCodec(ir codectypes.InterfaceRegistry, cdc *codec.ProtoCodec) OptionFunc {
	return func(o *options) {
		o.interfaceRegistry = ir
		o.cdc = cdc
	}
}

// Client implements a single network client to interact with cosmos based chains
type Client struct {
	auth auth.QueryClient
	bank bank.QueryClient

	ir codectypes.InterfaceRegistry

	clientCtx client.Context
}

// NewSingle instantiates a single network client
func NewSingle(grpcEndpoint, tendermintEndpoint string, optsFunc ...OptionFunc) (*Client, error) {
	opts := newDefaultOptions()
	for _, optFunc := range optsFunc {
		optFunc(&opts)
	}

	grpcConn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	tmRPC, err := http.New(tendermintEndpoint, tmWebsocketPath)
	if err != nil {
		return nil, err
	}

	authClient := auth.NewQueryClient(grpcConn)
	bankClient := bank.NewQueryClient(grpcConn)

	// NodeURI and Client are set from here otherwise
	// WitNodeURI will require to create a new client
	// it's done here because WithNodeURI panics if
	// connection to tendermint node fails
	clientCtx := client.Context{
		Client:  tmRPC,
		NodeURI: tendermintEndpoint,
	}
	clientCtx = clientCtx.
		WithJSONMarshaler(opts.cdc).
		WithInterfaceRegistry(opts.interfaceRegistry).
		WithTxConfig(tx.NewTxConfig(opts.cdc, tx.DefaultSignModes)).
		WithAccountRetriever(auth.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock)

	return &Client{
		auth:      authClient,
		bank:      bankClient,
		clientCtx: clientCtx,
		ir:        opts.interfaceRegistry,
	}, nil
}

func (c *Client) AccountInfo(ctx context.Context, addr string, height *int64) (auth.AccountI, error) {
	if height != nil {
		strHeight := strconv.FormatInt(*height, 10)
		ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	}

	accountInfo, err := c.auth.Account(ctx, &auth.QueryAccountRequest{
		Address: addr,
	})
	if err != nil {
		return nil, rosetta.FromGRPCToRosettaError(err)
	}

	var account auth.AccountI
	err = c.ir.UnpackAny(accountInfo.Account, &account)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrCodec, err.Error())
	}

	return account, nil
}

func (c *Client) Balances(ctx context.Context, addr string, height *int64) ([]sdk.Coin, error) {
	if height != nil {
		strHeight := strconv.FormatInt(*height, 10)
		ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	}
	balance, err := c.bank.AllBalances(ctx, &bank.QueryAllBalancesRequest{
		Address: addr,
	})
	if err != nil {
		return nil, rosetta.FromGRPCToRosettaError(err)
	}
	return balance.Balances, nil
}

// BlockByHash returns the block and the transactions contained in it given its height
func (c *Client) BlockByHash(ctx context.Context, hash string) (*tmtypes.ResultBlock, []*rosetta.SdkTxWithHash, error) {
	bHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, nil, rosetta.WrapError(rosetta.ErrBadArgument, fmt.Sprintf("invalid block hash: %s", err))
	}

	block, err := c.clientCtx.Client.BlockByHash(ctx, bHash)
	if err != nil {
		return nil, nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error()) // can be either a connection error or bad argument?
	}

	txs, err := c.ListTransactionsInBlock(ctx, block.Block.Height)
	if err != nil {
		return nil, nil, err
	}
	return block, txs, nil
}

// BlockByHeight returns the block and the transactions contained inside it given its height
func (c *Client) BlockByHeight(ctx context.Context, height *int64) (*tmtypes.ResultBlock, []*rosetta.SdkTxWithHash, error) {
	block, err := c.clientCtx.Client.Block(ctx, height)
	if err != nil {
		return nil, nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	txs, err := c.ListTransactionsInBlock(ctx, block.Block.Height)
	if err != nil {
		return nil, nil, err
	}
	return block, txs, err
}

// Coins fetches the existing coins in the application
func (c *Client) Coins(ctx context.Context) (sdk.Coins, error) {
	supply, err := c.bank.TotalSupply(ctx, &bank.QueryTotalSupplyRequest{})
	if err != nil {
		return nil, rosetta.FromGRPCToRosettaError(err)
	}
	return supply.Supply, nil
}

// ListTransactionsInBlock returns the list of the transactions in a block given its height
func (c *Client) ListTransactionsInBlock(ctx context.Context, height int64) ([]*rosetta.SdkTxWithHash, error) {
	txQuery := fmt.Sprintf(`tx.height=%d`, height)
	txList, err := c.clientCtx.Client.TxSearch(ctx, txQuery, true, nil, nil, "")
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}

	sdkTxs, err := tmResultTxsToSdkTxsWithHash(c.clientCtx.TxConfig.TxDecoder(), txList.Txs)
	if err != nil {
		return nil, err
	}
	return sdkTxs, nil
}

// GetTx returns a transaction given its hash
func (c *Client) GetTx(_ context.Context, hash string) (sdk.Tx, error) {
	txResp, err := authclient.QueryTx(c.clientCtx, hash)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	var sdkTx sdk.Tx
	err = c.ir.UnpackAny(txResp.Tx, &sdkTx)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrCodec, err.Error())
	}
	return sdkTx, nil
}

// GetUnconfirmedTx gets an unconfirmed transaction given its hash
// NOTE(fdymylja): not implemented yet
func (c *Client) GetUnconfirmedTx(_ context.Context, _ string) (sdk.Tx, error) {
	return nil, rosetta.WrapError(rosetta.ErrNotImplemented, "get unconfirmed transaction method is not supported")
}

// Mempool returns the unconfirmed transactions in the mempool
func (c *Client) Mempool(ctx context.Context) (*tmtypes.ResultUnconfirmedTxs, error) {
	txs, err := c.clientCtx.Client.UnconfirmedTxs(ctx, nil)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	return txs, nil
}

// Peers gets the number of peers
func (c *Client) Peers(ctx context.Context) ([]tmtypes.Peer, error) {
	netInfo, err := c.clientCtx.Client.NetInfo(ctx)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	return netInfo.Peers, nil
}

func (c *Client) Status(ctx context.Context) (*tmtypes.ResultStatus, error) {
	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	return status, err
}

func (c *Client) GetTxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}

func (c *Client) PostTx(txBytes []byte) (res *sdk.TxResponse, err error) {
	return c.clientCtx.BroadcastTx(txBytes)
}
