package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/cosmos/conversion"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"
)

// interface assertion
var _ rosetta.DataAPIClient = (*Client)(nil)

const tmWebsocketPath = "/websocket"

// options defines optional settings for SingleClient
type options struct {
	interfaceRegistry codectypes.InterfaceRegistry
	cdc               *codec.ProtoCodec
}

// newDefaultOptions builds the default options
func newDefaultOptions() options {
	// create codec and interface registry
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

	client client.Context
}

// NewSingleNetwork instantiates a single network client
func NewSingle(grpcEndpoint, tendermintEndpoint string, optsFunc ...OptionFunc) (*Client, error) {
	// set options
	opts := newDefaultOptions()
	for _, optFunc := range optsFunc {
		optFunc(&opts)
	}
	// connect to gRPC endpoint
	grpcConn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	// connect to tendermint
	tmRPC, err := http.New(tendermintEndpoint, tmWebsocketPath)
	if err != nil {
		return nil, err
	}
	// build clients
	authClient := auth.NewQueryClient(grpcConn)
	bankClient := bank.NewQueryClient(grpcConn)
	// build client context
	// NodeURI and Client are set from here otherwise
	// WitNodeURI will require to create a new client
	// it's done here because WithNodeURI panics if
	// connection to tendermint node fails
	clientContext := client.Context{
		Client:  tmRPC,
		NodeURI: tendermintEndpoint,
	}
	clientContext = clientContext.
		WithJSONMarshaler(opts.cdc).
		WithInterfaceRegistry(opts.interfaceRegistry).
		WithTxConfig(tx.NewTxConfig(opts.cdc, tx.DefaultSignModes)).
		WithAccountRetriever(auth.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock)

	// done
	return &Client{
		auth:   authClient,
		bank:   bankClient,
		client: clientContext,
		ir:     opts.interfaceRegistry,
	}, nil
}

func (c *Client) Balances(ctx context.Context, addr string, height *int64) ([]sdk.Coin, error) {
	// if height is set, send height instruction to account
	if height != nil {
		ctx = context.WithValue(ctx, grpctypes.GRPCBlockHeightHeader, *height)
	}
	// retrieve balance
	balance, err := c.bank.AllBalances(ctx, &bank.QueryAllBalancesRequest{
		Address: addr,
	})
	if err != nil {
		return nil, rosetta.FromGRPCToRosettaError(err)
	}
	// success
	return balance.Balances, nil
}

// BlockByHash returns the block and the transactions contained in it given its height
func (c *Client) BlockByHash(ctx context.Context, hash string) (*tmtypes.ResultBlock, []*rosetta.SdkTxWithHash, error) {
	bHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, nil, rosetta.WrapError(rosetta.ErrBadArgument, fmt.Sprintf("invalid block hash: %s", err))
	}
	// get block
	block, err := c.client.Client.BlockByHash(ctx, bHash)
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
	block, err := c.client.Client.Block(ctx, height)
	if err != nil {
		return nil, nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	txs, err := c.ListTransactionsInBlock(ctx, block.Block.Height)
	if err != nil {
		return nil, nil, err
	}
	return block, txs, err
}

// ListTransactionsInBlock returns the list of the transactions in a block given its height
func (c *Client) ListTransactionsInBlock(ctx context.Context, height int64) ([]*rosetta.SdkTxWithHash, error) {
	// prepare tx list query
	txQuery := fmt.Sprintf(`tx.height=%d`, height)
	// get tx
	txList, err := c.client.Client.TxSearch(ctx, txQuery, true, nil, nil, "")
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	sdkTxs, err := conversion.TmResultTxsToSdkTxs(c.client.TxConfig.TxDecoder(), txList.Txs)
	return sdkTxs, nil
}

// GetTx returns a transaction given its hash
func (c *Client) GetTx(_ context.Context, hash string) (sdk.Tx, error) {
	txResp, err := authclient.QueryTx(c.client, hash)
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
func (c *Client) GetUnconfirmedTx(_ context.Context, hash string) (sdk.Tx, error) {
	return nil, rosetta.WrapError(rosetta.ErrNotImplemented, "get unconfirmed transaction method is not supported")
}

// Mempool returns the unconfirmed transactions in the mempool
func (c *Client) Mempool(ctx context.Context) (*tmtypes.ResultUnconfirmedTxs, error) {
	txs, err := c.client.Client.UnconfirmedTxs(ctx, nil)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	return txs, nil
}

// Peers gets the number of peers
func (c *Client) Peers(ctx context.Context) ([]tmtypes.Peer, error) {
	netInfo, err := c.client.Client.NetInfo(ctx)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	return netInfo.Peers, nil
}

func (c *Client) Status(ctx context.Context) (*tmtypes.ResultStatus, error) {
	status, err := c.client.Client.Status(ctx)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	return status, err
}
