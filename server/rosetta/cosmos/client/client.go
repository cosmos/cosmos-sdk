package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/rosetta-sdk-go/types"
	"google.golang.org/grpc/metadata"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/cosmos/conversion"
	"github.com/cosmos/cosmos-sdk/server/rosetta/services"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
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

func (c *Client) SupportedOperations() []string {
	var supportedOperations []string
	for _, ii := range c.ir.ListImplementations("cosmos.base.v1beta1.Msg") {
		resolve, err := c.ir.Resolve(ii)
		if err == nil {
			if _, ok := resolve.(rosetta.Msg); ok {
				supportedOperations = append(supportedOperations, strings.TrimLeft(ii, "/"))
			}
		}
	}

	supportedOperations = append(supportedOperations, rosetta.OperationFee)

	return supportedOperations
}

func (c *Client) PreprocessOperationsToOptions(ctx context.Context, req *types.ConstructionPreprocessRequest) (options map[string]interface{}, err error) {
	operations := req.Operations
	if len(operations) > 3 {
		return nil, rosetta.ErrInvalidRequest
	}

	msgs, err := conversion.ConvertOpsToMsgs(c.ir, operations)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidAddress, err.Error())
	}

	if len(msgs) < 1 || len(msgs[0].GetSigners()) < 1 {
		return nil, rosetta.WrapError(rosetta.ErrInterpreting, "invalid msgs from operations")
	}

	memo, ok := req.Metadata["memo"]
	if !ok {
		memo = ""
	}

	defaultGas := float64(200000)

	gas := req.SuggestedFeeMultiplier
	if gas == nil {
		gas = &defaultGas
	}

	return map[string]interface{}{
		rosetta.OptionAddress: msgs[0].GetSigners()[0],
		rosetta.OptionMemo:    memo,
		rosetta.OptionGas:     gas,
	}, nil
}

func (c *Client) ConstructionPayload(ctx context.Context, request *types.ConstructionPayloadsRequest) (resp *types.ConstructionPayloadsResponse, err error) {
	if len(request.Operations) > 3 {
		return nil, rosetta.ErrInvalidOperation
	}

	msgs, fee, err := conversion.RosettaOperationsToSdkMsg(c.ir, request.Operations)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidOperation, err.Error())
	}

	metadata, err := services.GetMetadataFromPayloadReq(request)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrInvalidRequest, err.Error())
	}

	txFactory := tx.Factory{}.WithAccountNumber(metadata.AccountNumber).WithChainID(metadata.ChainID).
		WithGas(metadata.Gas).WithSequence(metadata.Sequence).WithMemo(metadata.Memo).WithFees(fee.String())

	TxConfig := c.getTxConfig()
	txFactory = txFactory.WithTxConfig(TxConfig)

	txBldr, err := tx.BuildUnsignedTx(txFactory, msgs...)
	if err != nil {
		return nil, err
	}

	if txFactory.SignMode() == signing.SignMode_SIGN_MODE_UNSPECIFIED {
		txFactory = txFactory.WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	}

	signerData := authsigning.SignerData{
		ChainID:       txFactory.ChainID(),
		AccountNumber: txFactory.AccountNumber(),
		Sequence:      txFactory.Sequence(),
	}

	signBytes, err := TxConfig.SignModeHandler().GetSignBytes(txFactory.SignMode(), signerData, txBldr.GetTx())
	if err != nil {
		return nil, err
	}

	txBytes, err := TxConfig.TxEncoder()(txBldr.GetTx())
	if err != nil {
		return nil, err
	}

	accIdentifiers := getAccountIdentifiersByMsgs(msgs)

	payloads := make([]*types.SigningPayload, len(accIdentifiers))
	for i, accID := range accIdentifiers {
		payloads[i] = &types.SigningPayload{
			AccountIdentifier: accID,
			Bytes:             crypto.Sha256(signBytes),
			SignatureType:     "ecdsa",
		}
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: hex.EncodeToString(txBytes),
		Payloads:            payloads,
	}, nil
}

func getAccountIdentifiersByMsgs(msgs []sdk.Msg) []*types.AccountIdentifier {
	var accIdentifiers []*types.AccountIdentifier
	for _, msg := range msgs {
		for _, signer := range msg.GetSigners() {
			accIdentifiers = append(accIdentifiers, &types.AccountIdentifier{Address: signer.String()})
		}
	}

	return accIdentifiers
}

func (c *Client) ConstructionMetadataFromOptions(ctx context.Context, options map[string]interface{}) (meta map[string]interface{}, err error) {
	if len(options) == 0 {
		return nil, rosetta.ErrInterpreting
	}

	addr, ok := options[rosetta.OptionAddress]
	if !ok {
		return nil, rosetta.ErrInvalidAddress
	}

	addrString := addr.(string)

	accountInfo, err := c.accountInfo(ctx, addrString, nil)
	if err != nil {
		return nil, err
	}

	gas, ok := options[rosetta.OptionGas]
	if !ok {
		return nil, rosetta.WrapError(rosetta.ErrInvalidAddress, "gas not set")
	}

	memo, ok := options[rosetta.OptionMemo]
	if !ok {
		return nil, rosetta.WrapError(rosetta.ErrInvalidMemo, "memo not set")
	}

	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		rosetta.AccountNumber: accountInfo.GetAccountNumber(),
		rosetta.Sequence:      accountInfo.GetSequence(),
		rosetta.ChainID:       status.NodeInfo.Network,
		rosetta.OptionGas:     gas,
		rosetta.OptionMemo:    memo,
	}, nil
}

// NewSingle instantiates a single network client
func NewSingle(grpcEndpoint, tendermintEndpoint string, optsFunc ...OptionFunc) (rosetta.NodeClient, error) {
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
		WithTxConfig(authtx.NewTxConfig(opts.cdc, authtx.DefaultSignModes)).
		WithAccountRetriever(auth.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock)

	return &Client{
		auth:      authClient,
		bank:      bankClient,
		clientCtx: clientCtx,
		ir:        opts.interfaceRegistry,
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
		return nil, rosetta.FromGRPCToRosettaError(err)
	}

	var account auth.AccountI
	err = c.ir.UnpackAny(accountInfo.Account, &account)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrCodec, err.Error())
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
		return nil, rosetta.FromGRPCToRosettaError(err)
	}

	availableCoins, err := c.coins(ctx)
	if err != nil {
		return nil, err
	}

	return conversion.SdkCoinsToRosettaAmounts(balance.Balances, availableCoins), nil
}

func (c *Client) BlockByHash(ctx context.Context, hash string) (rosetta.BlockResponse, error) {
	bHash, err := hex.DecodeString(hash)
	if err != nil {
		return rosetta.BlockResponse{}, fmt.Errorf("invalid block hash: %s", err)
	}

	block, err := c.clientCtx.Client.BlockByHash(ctx, bHash)
	if err != nil {
		return rosetta.BlockResponse{}, err
	}

	return buildBlockResponse(block), nil
}

func (c *Client) BlockByHeight(ctx context.Context, height *int64) (rosetta.BlockResponse, error) {
	block, err := c.clientCtx.Client.Block(ctx, height)
	if err != nil {
		return rosetta.BlockResponse{}, err
	}

	return buildBlockResponse(block), nil
}

func buildBlockResponse(block *tmtypes.ResultBlock) rosetta.BlockResponse {
	return rosetta.BlockResponse{
		Block:                conversion.TMBlockToRosettaBlockIdentifier(block),
		ParentBlock:          conversion.TMBlockToRosettaParentBlockIdentifier(block),
		MillisecondTimestamp: conversion.TimeToMilliseconds(block.Block.Time),
		TxCount:              int64(len(block.Block.Txs)),
	}
}

func (c *Client) BlockTransactionsByHash(ctx context.Context, hash string) (rosetta.BlockTransactionsResponse, error) {
	blockResp, err := c.BlockByHash(ctx, hash)
	if err != nil {
		return rosetta.BlockTransactionsResponse{}, err
	}

	txs, err := c.ListTransactionsInBlock(ctx, blockResp.Block.Index)
	if err != nil {
		return rosetta.BlockTransactionsResponse{}, err
	}

	return rosetta.BlockTransactionsResponse{
		BlockResponse: blockResp,
		Transactions:  conversion.SdkTxsWithHashToRosettaTxs(txs),
	}, nil
}

func (c *Client) BlockTransactionsByHeight(ctx context.Context, height *int64) (rosetta.BlockTransactionsResponse, error) {
	blockResp, err := c.BlockByHeight(ctx, height)
	if err != nil {
		return rosetta.BlockTransactionsResponse{}, err
	}

	txs, err := c.ListTransactionsInBlock(ctx, blockResp.Block.Index)
	if err != nil {
		return rosetta.BlockTransactionsResponse{}, err
	}

	return rosetta.BlockTransactionsResponse{
		BlockResponse: blockResp,
		Transactions:  conversion.SdkTxsWithHashToRosettaTxs(txs),
	}, nil
}

// Coins fetches the existing coins in the application
func (c *Client) coins(ctx context.Context) (sdk.Coins, error) {
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

	return conversion.SdkTxToOperations(txBldr.GetTx(), false, false), accountIdentifierSigners, nil
}

// GetTx returns a transaction given its hash
func (c *Client) GetTx(_ context.Context, hash string) (*types.Transaction, error) {
	txResp, err := authclient.QueryTx(c.clientCtx, hash)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	var sdkTx sdk.Tx
	err = c.ir.UnpackAny(txResp.Tx, &sdkTx)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrCodec, err.Error())
	}
	return conversion.SdkTxWithHashToRosettaTx(&rosetta.SdkTxWithHash{
		HexHash: txResp.TxHash,
		Code:    txResp.Code,
		Log:     txResp.RawLog,
		Tx:      sdkTx,
	}), nil
}

// GetUnconfirmedTx gets an unconfirmed transaction given its hash
// NOTE(fdymylja): not implemented yet
func (c *Client) GetUnconfirmedTx(_ context.Context, _ string) (*types.Transaction, error) {
	return nil, rosetta.WrapError(rosetta.ErrNotImplemented, "get unconfirmed transaction method is not supported")
}

// Mempool returns the unconfirmed transactions in the mempool
func (c *Client) Mempool(ctx context.Context) ([]*types.TransactionIdentifier, error) {
	txs, err := c.clientCtx.Client.UnconfirmedTxs(ctx, nil)
	if err != nil {
		return nil, err
	}

	return conversion.TMTxsToRosettaTxsIdentifiers(txs.Txs), nil
}

// Peers gets the number of peers
func (c *Client) Peers(ctx context.Context) ([]*types.Peer, error) {
	netInfo, err := c.clientCtx.Client.NetInfo(ctx)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	return conversion.TmPeersToRosettaPeers(netInfo.Peers), nil
}

func (c *Client) Status(ctx context.Context) (*types.SyncStatus, error) {
	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, rosetta.WrapError(rosetta.ErrUnknown, err.Error())
	}
	return conversion.TMStatusToRosettaSyncStatus(status), err
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
			rosetta.Log: res.RawLog,
		}, nil
}

func (c *Client) SignedTx(ctx context.Context, txBytes []byte, signatures []*types.Signature) (signedTxBytes []byte, err error) {
	TxConfig := c.getTxConfig()
	rawTx, err := TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, err
	}

	txBldr, _ := TxConfig.WrapTxBuilder(rawTx)

	var sigs = make([]signing.SignatureV2, len(signatures))
	for i, signature := range signatures {
		if signature.PublicKey.CurveType != "secp256k1" {
			return nil, rosetta.ErrUnsupportedCurve
		}

		cmp, err := btcec.ParsePubKey(signature.PublicKey.Bytes, btcec.S256())
		if err != nil {
			return nil, err
		}

		compressedPublicKey := make([]byte, secp256k1.PubKeySize)
		copy(compressedPublicKey, cmp.SerializeCompressed())
		pubKey := &secp256k1.PubKey{Key: compressedPublicKey}

		accountInfo, err := c.accountInfo(ctx, sdk.AccAddress(pubKey.Address()).String(), nil)
		if err != nil {
			return nil, err
		}

		sig := signing.SignatureV2{
			PubKey: pubKey,
			Data: &signing.SingleSignatureData{
				SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
				Signature: signature.Bytes,
			},
			Sequence: accountInfo.GetSequence(),
		}
		sigs[i] = sig
	}

	if err = txBldr.SetSignatures(sigs...); err != nil {
		return nil, err
	}

	txBytes, err = c.getTxConfig().TxEncoder()(txBldr.GetTx())
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}
