package systemtests

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	signing "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// LoadTestBroadcaster signs and broadcasts bank send txs in-process (no simd subprocess).
// Use for load tests to avoid spawning 10k simd processes.
// Call Close() when done to release the gRPC connection.
type LoadTestBroadcaster struct {
	clientCtx       client.Context
	grpcConn        *grpc.ClientConn
	accountNumsMu   sync.RWMutex
	accountNumsByID map[string]uint64
}

// NewLoadTestBroadcaster creates a broadcaster that uses the keyring at keyringDir,
// broadcasts to the RPC node at nodeAddr, and queries accounts via gRPC at grpcAddr.
func NewLoadTestBroadcaster(keyringDir, chainID, nodeAddr, grpcAddr string) (*LoadTestBroadcaster, error) {
	encCfg := testutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, keyringDir, nil, encCfg.Codec)
	if err != nil {
		return nil, fmt.Errorf("keyring: %w", err)
	}
	rpcClient, err := client.NewClientFromNode(nodeAddr)
	if err != nil {
		return nil, fmt.Errorf("rpc client: %w", err)
	}
	grpcConn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc: %w", err)
	}
	clientCtx := client.Context{}.
		WithCodec(encCfg.Codec).
		WithInterfaceRegistry(encCfg.InterfaceRegistry).
		WithTxConfig(encCfg.TxConfig).
		WithKeyring(kb).
		WithChainID(chainID).
		WithClient(rpcClient).
		WithGRPCClient(grpcConn).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode("sync").
		WithCmdContext(context.Background())
	return &LoadTestBroadcaster{
		clientCtx:       clientCtx,
		grpcConn:        grpcConn,
		accountNumsByID: make(map[string]uint64),
	}, nil
}

// Close releases the underlying gRPC connection.
func (b *LoadTestBroadcaster) Close() error {
	if b.grpcConn == nil {
		return nil
	}
	return b.grpcConn.Close()
}

// BroadcastBankSend signs and broadcasts a bank send tx. Returns tx hash and code (0 = success).
func (b *LoadTestBroadcaster) BroadcastBankSend(fromKey, toAddr, amount, fees string) (txHash string, code uint32, err error) {
	return b.broadcastBankSend(fromKey, toAddr, amount, fees, false, 0)
}

// BroadcastBankSendUnordered signs and broadcasts an unordered bank send tx.
// uniqueIdx ensures each tx has a distinct timeout (required for unordered nonce).
func (b *LoadTestBroadcaster) BroadcastBankSendUnordered(fromKey, toAddr, amount, fees string, uniqueIdx int) (txHash string, code uint32, err error) {
	return b.broadcastBankSend(fromKey, toAddr, amount, fees, true, uniqueIdx)
}

func (b *LoadTestBroadcaster) broadcastBankSend(fromKey, toAddr, amount, fees string, unordered bool, uniqueIdx int) (txHash string, code uint32, err error) {
	k, err := b.clientCtx.Keyring.Key(fromKey)
	if err != nil {
		return "", 0, fmt.Errorf("key %s: %w", fromKey, err)
	}
	pubKey, err := k.GetPubKey()
	if err != nil {
		return "", 0, err
	}
	fromAddr := sdk.AccAddress(pubKey.Address())

	toAcc, err := sdk.AccAddressFromBech32(toAddr)
	if err != nil {
		return "", 0, fmt.Errorf("to address: %w", err)
	}
	coins, err := sdk.ParseCoinsNormalized(amount)
	if err != nil {
		return "", 0, fmt.Errorf("amount: %w", err)
	}
	feeCoins, err := sdk.ParseCoinsNormalized(fees)
	if err != nil {
		return "", 0, fmt.Errorf("fees: %w", err)
	}
	msg := &banktypes.MsgSend{
		FromAddress: fromAddr.String(),
		ToAddress:   toAcc.String(),
		Amount:      coins,
	}
	txBuilder := b.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return "", 0, err
	}
	txBuilder.SetFeeAmount(feeCoins)
	txBuilder.SetGasLimit(200_000)
	txBuilder.SetMemo("load-test")
	if unordered {
		txBuilder.SetUnordered(true)
		// Fixed base + ms-scale offset guarantees unique timeouts across concurrent goroutines.
		base := time.Now().Add(5 * time.Minute)
		txBuilder.SetTimeoutTimestamp(base.Add(time.Duration(uniqueIdx) * time.Millisecond))
	}

	txFactory := clienttx.Factory{}.
		WithTxConfig(b.clientCtx.TxConfig).
		WithChainID(b.clientCtx.ChainID).
		WithKeybase(b.clientCtx.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)
	if unordered {
		accNum, err := b.getCachedAccountNumber(fromKey, fromAddr)
		if err != nil {
			return "", 0, err
		}
		txFactory = txFactory.WithAccountNumber(accNum).WithSequence(0)
	} else {
		accNum, seq, err := b.clientCtx.AccountRetriever.GetAccountNumberSequence(b.clientCtx, fromAddr)
		if err != nil {
			return "", 0, err
		}
		txFactory = txFactory.WithAccountNumber(accNum).WithSequence(seq)
	}
	if err := authclient.SignTx(txFactory, b.clientCtx, fromKey, txBuilder, unordered, true); err != nil {
		return "", 0, err
	}
	txBytes, err := b.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", 0, err
	}
	res, err := b.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return "", 0, err
	}
	return res.TxHash, res.Code, nil
}

func (b *LoadTestBroadcaster) getCachedAccountNumber(fromKey string, fromAddr sdk.AccAddress) (uint64, error) {
	b.accountNumsMu.RLock()
	accNum, ok := b.accountNumsByID[fromKey]
	b.accountNumsMu.RUnlock()
	if ok {
		return accNum, nil
	}

	accNum, _, err := b.clientCtx.AccountRetriever.GetAccountNumberSequence(b.clientCtx, fromAddr)
	if err != nil {
		return 0, err
	}

	b.accountNumsMu.Lock()
	if existing, found := b.accountNumsByID[fromKey]; found {
		b.accountNumsMu.Unlock()
		return existing, nil
	}
	b.accountNumsByID[fromKey] = accNum
	b.accountNumsMu.Unlock()
	return accNum, nil
}

// KeyringDir returns the keyring directory for the given work dir and output dir.
func KeyringDir(workDir, outputDir string) string {
	return filepath.Join(workDir, outputDir)
}
