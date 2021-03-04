package rosetta

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/version"

	abcitypes "github.com/tendermint/tendermint/abci/types"

	rosettatypes "github.com/coinbase/rosetta-sdk-go/types"
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

	converter Converter
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
		converter: NewConverter(cfg.Codec, cfg.InterfaceRegistry, authtx.NewTxConfig(cfg.Codec, authtx.DefaultSignModes)),
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

func (c *Client) Balances(ctx context.Context, addr string, height *int64) ([]*rosettatypes.Amount, error) {
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

	return c.converter.ToRosetta().CoinsToAmounts(balance.Balances, availableCoins), nil
}

func (c *Client) BlockByHash(ctx context.Context, hash string) (crgtypes.BlockResponse, error) {
	bHash, err := hex.DecodeString(hash)
	if err != nil {
		return crgtypes.BlockResponse{}, fmt.Errorf("invalid block hash: %s", err)
	}

	block, err := c.clientCtx.Client.BlockByHash(ctx, bHash)
	if err != nil {
		return crgtypes.BlockResponse{}, crgerrs.WrapError(crgerrs.ErrBadGateway, err.Error())
	}

	return c.converter.ToRosetta().BlockResponse(block), nil
}

func (c *Client) BlockByHeight(ctx context.Context, height *int64) (crgtypes.BlockResponse, error) {
	block, err := c.clientCtx.Client.Block(ctx, height)
	if err != nil {
		return crgtypes.BlockResponse{}, crgerrs.WrapError(crgerrs.ErrBadGateway, err.Error())
	}

	return c.converter.ToRosetta().BlockResponse(block), nil
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
	bb, _ := json.Marshal(blockTxResp)
	log.Printf("block %d: %s", height, bb)
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

func (c *Client) TxOperationsAndSignersAccountIdentifiers(signed bool, txBytes []byte) (ops []*rosettatypes.Operation, signers []*rosettatypes.AccountIdentifier, err error) {
	txConfig := c.getTxConfig()
	rawTx, err := c.txDecoder(txBytes)
	if err != nil {
		return nil, nil, err
	}

	txBldr, err := txConfig.WrapTxBuilder(rawTx)
	if err != nil {
		return nil, nil, err
	}

	var accountIdentifierSigners []*rosettatypes.AccountIdentifier
	if signed {
		addrs := txBldr.GetTx().GetSigners()
		for _, addr := range addrs {
			signer := &rosettatypes.AccountIdentifier{
				Address: addr.String(),
			}
			accountIdentifierSigners = append(accountIdentifierSigners, signer)
		}
	}

	rosTx, err := c.converter.ToRosetta().Tx(txBytes, nil)
	if err != nil {
		return nil, nil, err
	}

	return rosTx.Operations, accountIdentifierSigners, nil
}

// GetTx returns a transaction given its hash, in rosetta begin block and end block are mocked
// as transaction hashes in order to adhere to balance tracking rules
func (c *Client) GetTx(ctx context.Context, hash string) (*rosettatypes.Transaction, error) {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, fmt.Sprintf("bad tx hash: %s", err))
	}

	// get tx type and hash
	txType, hashBytes := c.converter.FromRosetta().HashToTxType(hashBytes)

	// construct rosetta tx
	switch txType {
	// handle begin block hash
	case BeginBlockTx:
		// get block height by hash
		block, err := c.clientCtx.Client.BlockByHash(ctx, hashBytes)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
		}

		// get block txs
		fullBlock, err := c.blockTxs(ctx, &block.Block.Height)
		if err != nil {
			return nil, err
		}

		return fullBlock.Transactions[0], nil
	// handle deliver tx hash
	case DeliverTxTx:
		rawTx, err := c.clientCtx.Client.Tx(ctx, hashBytes, true)
		if err != nil {
			log.Printf("tx err: %s : %s", hash, err)
			return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
		}
		return c.converter.ToRosetta().Tx(rawTx.Tx, &rawTx.TxResult)
	// handle end block hash
	case EndBlockTx:
		// get block height by hash
		block, err := c.clientCtx.Client.BlockByHash(ctx, hashBytes)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
		}

		// get block txs
		fullBlock, err := c.blockTxs(ctx, &block.Block.Height)
		if err != nil {
			return nil, err
		}

		// get last tx
		return fullBlock.Transactions[len(fullBlock.Transactions)-1], nil
	// unrecognized tx
	default:
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, fmt.Sprintf("invalid tx hash provided: %s", hash))
	}
}

// GetUnconfirmedTx gets an unconfirmed transaction given its hash
func (c *Client) GetUnconfirmedTx(ctx context.Context, hash string) (*rosettatypes.Transaction, error) {
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
	case BeginEndBlockTxSize:
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, fmt.Sprintf("endblock and begin block txs cannot be unconfirmed"))
	case DeliverTxSize:
		break
	}

	// iterate over unconfirmed txs to find the one with matching hash
	for _, unconfirmedTx := range res.Txs {
		if !bytes.Equal(unconfirmedTx.Hash(), hashAsBytes) {
			continue
		}

		return c.converter.ToRosetta().Tx(unconfirmedTx, nil)
	}
	return nil, crgerrs.WrapError(crgerrs.ErrNotFound, "transaction not found in mempool: "+hash)
}

// Mempool returns the unconfirmed transactions in the mempool
func (c *Client) Mempool(ctx context.Context) ([]*rosettatypes.TransactionIdentifier, error) {
	txs, err := c.clientCtx.Client.UnconfirmedTxs(ctx, nil)
	if err != nil {
		return nil, err
	}

	return c.converter.ToRosetta().TxIdentifiers(txs.Txs), nil
}

// Peers gets the number of peers
func (c *Client) Peers(ctx context.Context) ([]*rosettatypes.Peer, error) {
	netInfo, err := c.clientCtx.Client.NetInfo(ctx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	return c.converter.ToRosetta().Peers(netInfo.Peers), nil
}

func (c *Client) Status(ctx context.Context) (*rosettatypes.SyncStatus, error) {
	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	return c.converter.ToRosetta().StatusToSyncStatus(status), err
}

func (c *Client) getTxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}

func (c *Client) PostTx(txBytes []byte) (*rosettatypes.TransactionIdentifier, map[string]interface{}, error) {
	// sync ensures it will go through checkTx
	res, err := c.clientCtx.BroadcastTxSync(txBytes)
	if err != nil {
		return nil, nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	// check if tx was broadcast successfully
	if res.Code != abcitypes.CodeTypeOK {
		return nil, nil, crgerrs.WrapError(crgerrs.ErrUnknown, fmt.Sprintf("transaction broadcast failure: (%d) %s ", res.Code, res.RawLog))
	}

	return &rosettatypes.TransactionIdentifier{
			Hash: res.TxHash,
		},
		map[string]interface{}{
			Log: res.RawLog,
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

// construction endpoints

// ConstructionMetadataFromOptions builds the metadata given the options
func (c *Client) ConstructionMetadataFromOptions(ctx context.Context, options map[string]interface{}) (meta map[string]interface{}, err error) {
	if len(options) == 0 {
		return nil, crgerrs.ErrBadArgument
	}

	constructionOptions := new(PreprocessOperationsOptionsResponse)

	err = constructionOptions.FromMetadata(options)
	if err != nil {
		return nil, err
	}

	signersData := make([]*SignerData, len(constructionOptions.ExpectedSigners))

	for i, signer := range constructionOptions.ExpectedSigners {
		accountInfo, err := c.accountInfo(ctx, signer, nil)
		if err != nil {
			return nil, err
		}

		signersData[i] = &SignerData{
			AccountNumber: accountInfo.GetAccountNumber(),
			Sequence:      accountInfo.GetSequence(),
		}
	}

	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, err
	}

	metadataResp := ConstructionMetadata{
		ChainID:     status.NodeInfo.Network,
		SignersData: signersData,
		GasLimit:    constructionOptions.GasLimit,
		GasPrice:    constructionOptions.GasPrice,
		Memo:        constructionOptions.Memo,
	}

	return metadataResp.ToMetadata()
}

func (c *Client) blockTxs(ctx context.Context, height *int64) (crgtypes.BlockTransactionsResponse, error) {
	// get block info
	blockInfo, err := c.clientCtx.Client.Block(ctx, height)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}
	// get block events
	blockResults, err := c.clientCtx.Client.BlockResults(ctx, height)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}

	if len(blockResults.TxsResults) != len(blockInfo.Block.Txs) {
		// wtf?
		panic("block results transactions do now match block transactions")
	}
	// process begin and end block txs
	beginBlockTx := &rosettatypes.Transaction{
		TransactionIdentifier: &rosettatypes.TransactionIdentifier{Hash: c.converter.ToRosetta().BeginBlockTxHash(blockInfo.BlockID.Hash)},
		Operations: AddOperationIndexes(
			nil,
			c.converter.ToRosetta().EventsToBalanceOps(StatusTxSuccess, blockResults.BeginBlockEvents),
		),
	}

	endBlockTx := &rosettatypes.Transaction{
		TransactionIdentifier: &rosettatypes.TransactionIdentifier{Hash: c.converter.ToRosetta().EndBlockTxHash(blockInfo.BlockID.Hash)},
		Operations: AddOperationIndexes(
			nil,
			c.converter.ToRosetta().EventsToBalanceOps(StatusTxSuccess, blockResults.EndBlockEvents),
		),
	}

	deliverTx := make([]*rosettatypes.Transaction, len(blockInfo.Block.Txs))
	// process normal txs
	for i, tx := range blockInfo.Block.Txs {
		rosTx, err := c.converter.ToRosetta().Tx(tx, blockResults.TxsResults[i])
		if err != nil {
			return crgtypes.BlockTransactionsResponse{}, err
		}
		deliverTx[i] = rosTx
	}

	finalTxs := make([]*rosettatypes.Transaction, 0, 2+len(deliverTx))
	finalTxs = append(finalTxs, beginBlockTx)
	finalTxs = append(finalTxs, deliverTx...)
	finalTxs = append(finalTxs, endBlockTx)

	return crgtypes.BlockTransactionsResponse{
		BlockResponse: c.converter.ToRosetta().BlockResponse(blockInfo),
		Transactions:  finalTxs,
	}, nil
}
