package accounts

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/collections"
	aa_interface_v1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
	txdecode "cosmossdk.io/x/tx/decode"

	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// IsAbstractedAccount returns if the provided address is an abstracted account or not.
func (k Keeper) IsAbstractedAccount(ctx context.Context, addr []byte) (bool, error) {
	accType, err := k.AccountsByType.Get(ctx, addr)
	switch {
	case errors.Is(err, collections.ErrNotFound):
		return false, nil
	case err != nil:
		return false, err
	}

	impl, ok := k.accounts[accType]
	if !ok {
		return false, fmt.Errorf("%w: %s", errAccountTypeNotFound, accType)
	}
	return impl.HasExec(&aa_interface_v1.MsgAuthenticate{}), nil
}

// AuthenticateAccount runs the authentication flow of an account.
func (k Keeper) AuthenticateAccount(ctx context.Context, signer []byte, bundler string, rawTx *tx.TxRaw, protoTx *tx.Tx, signIndex uint32) error {
	msg := &aa_interface_v1.MsgAuthenticate{
		Bundler:     bundler,
		RawTx:       rawTx,
		Tx:          protoTx,
		SignerIndex: signIndex,
	}
	_, err := k.Execute(ctx, signer, address.Module("accounts"), msg, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAuthentication, err)
	}
	return nil
}

// ExecuteBundledTx will execute the single bundled tx.
func (k Keeper) ExecuteBundledTx(ctx context.Context, bundler string, txBytes []byte) *v1.BundledTxResponse {
	resp, err := k.executeBundledTx(ctx, bundler, txBytes)
	if err != nil {
		if resp == nil {
			return &v1.BundledTxResponse{
				Error: err.Error(),
			}
		}
		// ensure partial information is not discarded
		resp.Error = err.Error()
		return resp
	}
	return resp
}

func (k Keeper) executeBundledTx(ctx context.Context, bundler string, txBytes []byte) (*v1.BundledTxResponse, error) {
	bundledTx, err := k.txDecoder.Decode(txBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid tx bytes: %w", err)
	}
	blockInfo := k.HeaderService.HeaderInfo(ctx)
	xt, err := verifyAndExtractAaXtFromTx(bundledTx, uint64(blockInfo.Height), blockInfo.Time)
	if err != nil {
		return nil, fmt.Errorf("%w: tx failed validation check: %w", ErrAASemantics, err)
	}

	resp := new(v1.BundledTxResponse)
	// to execute a bundled tx the first step is authentication.
	signer := bundledTx.Signers[0]
	authGasUsed, err := k.BranchService.ExecuteWithGasLimit(ctx, xt.AuthenticationGasLimit, func(ctx context.Context) error {
		return k.AuthenticateAccount(ctx, signer, bundler, protov2TxRawToProtoV1(bundledTx.TxRaw), protoV2TxToProtoV1(bundledTx.Tx), 0)
	})
	resp.AuthenticationGasUsed = authGasUsed // set independently of outcome
	if err != nil {
		return resp, fmt.Errorf("%w: %w", ErrAuthentication, err)
	}

	// after authentication, we execute the bundler messages.
	if len(xt.BundlerPaymentMessages) != 0 {
		var paymentMsgResp []*implementation.Any
		bundlerPaymentGasUsed, err := k.BranchService.ExecuteWithGasLimit(ctx, xt.BundlerPaymentGasLimit, func(ctx context.Context) error {
			responses, err := k.sendAnyMessages(ctx, signer, xt.BundlerPaymentMessages)
			if err != nil {
				return err
			}
			paymentMsgResp = responses
			return nil
		})
		resp.BundlerPaymentGasUsed = bundlerPaymentGasUsed // set independently of outcome
		if err != nil {
			return resp, fmt.Errorf("%w: %w", ErrBundlerPayment, err)
		}
		resp.BundlerPaymentResponses = paymentMsgResp
	}

	// finally execute the real messages
	var execResponses []*implementation.Any
	execGasUsed, err := k.BranchService.ExecuteWithGasLimit(ctx, xt.ExecutionGasLimit, func(ctx context.Context) error {
		responses, err := k.sendManyMessagesReturnAnys(ctx, signer, bundledTx.Messages)
		if err != nil {
			return err
		}
		execResponses = responses
		return nil
	})
	resp.ExecutionGasUsed = execGasUsed // set independently of outcome
	if err != nil {
		return resp, fmt.Errorf("%w: %w", ErrExecution, err)
	}
	resp.ExecutionResponses = execResponses

	return resp, nil
}

var aaXtName = gogoproto.MessageName(&aa_interface_v1.TxExtension{})

func verifyAndExtractAaXtFromTx(bundledTx *txdecode.DecodedTx, currentBlock uint64, currentTime time.Time) (*aa_interface_v1.TxExtension, error) {
	// some basic things: we do not allow multi addresses in the bundled tx
	// rationale: the bundler could simply bundle multiple txs in the same bundle
	// with other accounts.
	if len(bundledTx.Signers) != 1 {
		return nil, fmt.Errorf("account abstraction bundled txs can only have one signer, got: %d", len(bundledTx.Signers))
	}
	// do not allow sign modes different from single
	if len(bundledTx.Tx.AuthInfo.SignerInfos) != 1 {
		return nil, fmt.Errorf("account abstraction tx must have one signer info")
	}

	// check sign mode is valid
	if bundledTx.Tx.AuthInfo.SignerInfos[0].ModeInfo.GetSingle() == nil {
		return nil, fmt.Errorf("account abstraction mode info must be single")
	}

	// we do not want the tx to have any fees set.
	if bundledTx.Tx.AuthInfo.Fee != nil {
		return nil, fmt.Errorf("account abstraction tx must not have the Fee field set")
	}

	// check timeouts TODO: do not like this much since it feels like we are adding repetition of logic.
	if bundledTx.Tx.Body.TimeoutTimestamp != nil && currentTime.After(bundledTx.Tx.Body.TimeoutTimestamp.AsTime()) {
		return nil, fmt.Errorf("block time is after tx timeout timestamp")
	}
	if bundledTx.Tx.Body.TimeoutHeight != 0 && currentBlock >= bundledTx.Tx.Body.TimeoutHeight {
		return nil, fmt.Errorf("block height is after tx timeout height")
	}

	// extract extension
	found := false
	xt := new(aa_interface_v1.TxExtension)
	for i, anyPb := range bundledTx.Tx.Body.ExtensionOptions {
		xtName := nameFromTypeURL(anyPb.TypeUrl)
		if xtName == aaXtName {
			if found {
				return nil, fmt.Errorf("multiple aa extensions on the same tx")
			}
			found = true
			// unwrap
			err := xt.Unmarshal(anyPb.Value)
			if err != nil {
				return nil, fmt.Errorf("unable to unmarshal tx extension at index %d: %w", i, err)
			}
		} else {
			log.Printf("name: %s, wanted: %s", xtName, aaXtName)
		}
	}
	if !found {
		return nil, fmt.Errorf("did not have AA extension %s", aaXtName)
	}

	err := verifyAaXt(xt)
	if err != nil {
		return nil, fmt.Errorf("invalid account abstraction tx extension: %w", err)
	}

	return xt, nil
}

func verifyAaXt(_ *aa_interface_v1.TxExtension) error {
	return nil
}

func nameFromTypeURL(url string) string {
	name := url
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		name = name[i+len("/"):]
	}
	return name
}
