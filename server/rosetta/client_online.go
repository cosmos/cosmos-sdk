package rosetta

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/version"

	abcitypes "github.com/tendermint/tendermint/abci/types"

	"github.com/coinbase/rosetta-sdk-go/types"
	"google.golang.org/grpc/metadata"

	"github.com/tendermint/tendermint/rpc/client/http"
	"google.golang.org/grpc"

	crgerrs "github.com/tendermint/cosmos-rosetta-gateway/errors"
	crgtypes "github.com/tendermint/cosmos-rosetta-gateway/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
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

	txDecoder sdk.TxDecoder
	version   string
}

// NewClient instantiates a new online servicer
func NewClient(cfg *Config) (*Client, error) {
	info := version.NewInfo()

	v := info.Version
	if v == "" {
		v = "unknown"
	}

	return &Client{
		config:    cfg,
		ir:        cfg.InterfaceRegistry,
		version:   fmt.Sprintf("%s/%s", info.AppName, v),
		txDecoder: authtx.NewTxConfig(cfg.Codec, authtx.DefaultSignModes).TxDecoder(),
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

	return tmResultBlockToRosettaBlockResponse(block), nil
}

func (c *Client) BlockByHeight(ctx context.Context, height *int64) (crgtypes.BlockResponse, error) {
	block, err := c.clientCtx.Client.Block(ctx, height)
	if err != nil {
		return crgtypes.BlockResponse{}, err
	}

	return tmResultBlockToRosettaBlockResponse(block), nil
}

func (c *Client) BlockTransactionsByHash(ctx context.Context, hash string) (crgtypes.BlockTransactionsResponse, error) {
	// TODO(fdymylja): use a faster path, by searching the block by hash, instead of doing a double query operation
	blockResp, err := c.BlockByHash(ctx, hash)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}

	return c.blockTxs(ctx, &blockResp.Block.Index)
}

func (c *Client) BlockTransactionsByHeight(ctx context.Context, height *int64) (crgtypes.BlockTransactionsResponse, error) {
	blockTxResp, err := c.blockTxs(ctx, height)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}

	return blockTxResp, nil
}

// Coins fetches the existing coins in the application
func (c *Client) coins(ctx context.Context) (sdk.Coins, error) {
	supply, err := c.bank.TotalSupply(ctx, &bank.QueryTotalSupplyRequest{})
	if err != nil {
		return nil, crgerrs.FromGRPCToRosettaError(err)
	}
	return supply.Supply, nil
}

func (c *Client) TxOperationsAndSignersAccountIdentifiers(signed bool, txBytes []byte) (ops []*types.Operation, signers []*types.AccountIdentifier, err error) {
	txConfig := c.getTxConfig()
	rawTx, err := c.txDecoder(txBytes)
	if err != nil {
		return nil, nil, err
	}

	txBldr, err := txConfig.WrapTxBuilder(rawTx)
	if err != nil {
		return nil, nil, err
	}

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
func (c *Client) GetTx(ctx context.Context, hash string) (*types.Transaction, error) {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, fmt.Sprintf("bad tx hash: %s", err))
	}

	// here we check for the hash length to understand if it is a begin or endblock tx or a standard tendermint tx
	switch len(hashBytes) {
	case beginEndBlockTxSize:
		// verify if it's end or begin block operations we're trying to query
		switch hashBytes[0] {
		case beginBlockHashStart:
			return c.beginBlockTx(ctx, hashBytes[1:])
		case endBlockHashStart:
			return c.endBlockTx(ctx, hashBytes[1:])
		default:
			return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, fmt.Sprintf("bad begin endblock starting byte: %x", hashBytes[0]))
		}
	// standard tx...
	case sha256.Size:
		return c.getTx(ctx, hashBytes)
	default:
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, fmt.Sprintf("invalid tx size: %d", len(hashBytes)))
	}
}

// GetUnconfirmedTx gets an unconfirmed transaction given its hash
func (c *Client) GetUnconfirmedTx(ctx context.Context, hash string) (*types.Transaction, error) {
	res, err := c.clientCtx.Client.UnconfirmedTxs(ctx, nil)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrNotFound, "unconfirmed tx not found")
	}

	hashAsBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrInterpreting, "invalid hash")
	}

	// assert that correct tx length is provided
	switch len(hashAsBytes) {
	default:
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, fmt.Sprintf("unrecognized tx size: %d", len(hashAsBytes)))
	case beginEndBlockTxSize:
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, fmt.Sprintf("endblock and begin block txs cannot be unconfirmed"))
	case deliverTxSize:
		break
	}

	// iterate over unconfirmed txs to find the one with matching hash
	for _, unconfirmedTx := range res.Txs {
		if !bytes.Equal(unconfirmedTx.Hash(), hashAsBytes) {
			continue
		}

		return sdkTxToRosettaTx(c.txDecoder, unconfirmedTx, nil)
	}
	return nil, crgerrs.WrapError(crgerrs.ErrNotFound, "transaction not found in mempool: "+hash)
}

// Mempool returns the unconfirmed transactions in the mempool
func (c *Client) Mempool(ctx context.Context) ([]*types.TransactionIdentifier, error) {
	txs, err := c.clientCtx.Client.UnconfirmedTxs(ctx, nil)
	if err != nil {
		return nil, err
	}

	return tmTxsToRosettaTxsIdentifiers(txs.Txs), nil
}

// Peers gets the number of peers
func (c *Client) Peers(ctx context.Context) ([]*types.Peer, error) {
	netInfo, err := c.clientCtx.Client.NetInfo(ctx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	return tmPeersToRosettaPeers(netInfo.Peers), nil
}

func (c *Client) Status(ctx context.Context) (*types.SyncStatus, error) {
	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	return tmStatusToRosettaSyncStatus(status), err
}

func (c *Client) getTxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}

func (c *Client) PostTx(txBytes []byte) (*types.TransactionIdentifier, map[string]interface{}, error) {
	// sync ensures it will go through checkTx
	res, err := c.clientCtx.BroadcastTxSync(txBytes)
	if err != nil {
		return nil, nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	// check if tx was broadcast successfully
	if res.Code != abcitypes.CodeTypeOK {
		return nil, nil, crgerrs.WrapError(crgerrs.ErrUnknown, fmt.Sprintf("transaction broadcast failure: (%d) %s ", res.Code, res.RawLog))
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

	addrString, ok := addr.(string)
	if !ok {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidAddress, "address is not a string")
	}

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
