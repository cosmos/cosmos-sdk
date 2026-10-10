package ante

import (
	"slices"
	"time"

	cmtmldsa65 "github.com/cometbft/cometbft/crypto/mldsa65"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// ValidateBasicDecorator will call tx.ValidateBasic and return any non-nil error.
// If ValidateBasic passes, decorator calls next AnteHandler in chain. Note,
// ValidateBasicDecorator decorator will not get executed on ReCheckTx since it
// is not dependent on application state.
type ValidateBasicDecorator struct{}

func NewValidateBasicDecorator() ValidateBasicDecorator {
	return ValidateBasicDecorator{}
}

func (vbd ValidateBasicDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// no need to validate basic on recheck tx, call next antehandler
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	if validateBasic, ok := tx.(sdk.HasValidateBasic); ok {
		if err := validateBasic.ValidateBasic(); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// ValidateMemoDecorator will validate memo given the parameters passed in
// If memo is too large decorator returns with error, otherwise call next AnteHandler
// CONTRACT: Tx must implement TxWithMemo interface
type ValidateMemoDecorator struct {
	ak AccountKeeper
}

func NewValidateMemoDecorator(ak AccountKeeper) ValidateMemoDecorator {
	return ValidateMemoDecorator{
		ak: ak,
	}
}

func (vmd ValidateMemoDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	memoLength := len(memoTx.GetMemo())
	if memoLength > 0 {
		params := vmd.ak.GetParams(ctx)
		if uint64(memoLength) > params.MaxMemoCharacters {
			return ctx, errorsmod.Wrapf(sdkerrors.ErrMemoTooLarge,
				"maximum number of characters is %d but received %d characters",
				params.MaxMemoCharacters, memoLength,
			)
		}
	}

	return next(ctx, tx, simulate)
}

// ConsumeTxSizeGasDecorator will take in parameters and consume gas proportional
// to the size of tx before calling next AnteHandler. Note, the gas costs will be
// slightly over estimated due to the fact that any given signing account may need
// to be retrieved from state.
//
// CONTRACT: If simulate=true, then signatures must either be completely filled
// in or empty.
// CONTRACT: To use this decorator, signatures of transaction must be represented
// as legacytx.StdSignature otherwise simulate mode will incorrectly estimate gas cost.
type ConsumeTxSizeGasDecorator struct {
	ak AccountKeeper
}

func NewConsumeGasForTxSizeDecorator(ak AccountKeeper) ConsumeTxSizeGasDecorator {
	return ConsumeTxSizeGasDecorator{
		ak: ak,
	}
}

func (cgts ConsumeTxSizeGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	params := cgts.ak.GetParams(ctx)

	ctx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*storetypes.Gas(len(ctx.TxBytes())), "txSize")

	// simulate gas cost for signatures in simulate mode
	if simulate {
		// in simulate mode, each element should be a nil signature
		sigs, err := sigTx.GetSignaturesV2()
		if err != nil {
			return ctx, err
		}
		n := len(sigs)

		signers, err := sigTx.GetSigners()
		if err != nil {
			return sdk.Context{}, err
		}

		for i, signer := range signers {
			// if signature is already filled in, no need to simulate gas cost
			if i < n && !isIncompleteSignature(sigs[i].Data) {
				continue
			}

			var pubkey cryptotypes.PubKey

			acc := cgts.ak.GetAccount(ctx, signer)

			// Size the placeholder from the stored pubkey. An account with none
			// yet (such as a native ML-DSA-65 account's first tx) falls back to
			// the SignerInfo pubkey, then to simSecp256k1Pubkey. This only
			// affects the simulate estimate.
			switch {
			case acc != nil && acc.GetPubKey() != nil:
				pubkey = acc.GetPubKey()
			case i < n && sigs[i].PubKey != nil:
				pubkey = sigs[i].PubKey
			default:
				pubkey = simSecp256k1Pubkey
			}

			ctx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*simSigTxSize(pubkey, params.TxSigLimit), "txSize")
		}
	}

	return next(ctx, tx, simulate)
}

// simSigTxSize estimates the bytes a missing signature for pubkey adds to a
// tx, for simulation only.
func simSigTxSize(pubkey cryptotypes.PubKey, txSigLimit uint64) storetypes.Gas {
	// use stdsignature to mock the size of a full signature
	stdSigSize := func(sigSize int) storetypes.Gas {
		simSig := legacytx.StdSignature{ //nolint:staticcheck // SA1019: legacytx.StdSignature is deprecated
			Signature: make([]byte, sigSize),
			PubKey:    pubkey,
		}
		return storetypes.Gas(len(legacy.Cdc.MustMarshal(simSig)) + 6)
	}

	cost := stdSigSize(len(simSecp256k1Sig))

	// If the pubkey is a multi-signature pubkey, then we estimate for the maximum
	// number of signers.
	if _, ok := pubkey.(*multisig.LegacyAminoPubKey); ok {
		cost *= txSigLimit
	}

	// Signatures larger than a secp256k1 one, such as ML-DSA-65 ones (possibly
	// from an account rekeyed with MsgChangePubKey), are sized for their key
	// type. This never lowers the estimate above.
	if sigSize := simSigSize(pubkey); sigSize > len(simSecp256k1Sig) {
		cost = max(cost, stdSigSize(sigSize))
	}

	return cost
}

// simSigSize returns an upper bound on the signature bytes pubkey produces: the
// fixed ML-DSA-65 signature size, the sum over all subkeys for a multisig, and
// the secp256k1 signature size for any other key. For a multisig, the summed
// size only exceeds the TxSigLimit multiplier in simSigTxSize when TxSigLimit
// is small (about 2 or less for ML-DSA-65 subkeys).
func simSigSize(pubkey cryptotypes.PubKey) int {
	switch pk := pubkey.(type) {
	case *mldsa65.PubKey:
		return cmtmldsa65.SignatureSize
	case *multisig.LegacyAminoPubKey:
		size := 0
		for _, sub := range pk.GetPubKeys() {
			size += simSigSize(sub)
		}
		return size
	default:
		return len(simSecp256k1Sig)
	}
}

// isIncompleteSignature tests whether SignatureData is fully filled in for simulation purposes
func isIncompleteSignature(data signing.SignatureData) bool {
	if data == nil {
		return true
	}

	switch data := data.(type) {
	case *signing.SingleSignatureData:
		return len(data.Signature) == 0
	case *signing.MultiSignatureData:
		if len(data.Signatures) == 0 {
			return true
		}
		if slices.ContainsFunc(data.Signatures, isIncompleteSignature) {
			return true
		}
	}

	return false
}

type (
	// TxTimeoutHeightDecorator defines an AnteHandler decorator that checks for a
	// tx height timeout.
	TxTimeoutHeightDecorator struct{}

	// TxWithTimeoutHeight defines the interface a tx must implement in order for
	// TxHeightTimeoutDecorator to process the tx.
	TxWithTimeoutHeight interface {
		sdk.Tx

		GetTimeoutHeight() uint64
		GetTimeoutTimeStamp() time.Time
	}
)

// NewTxTimeoutHeightDecorator defines an AnteHandler decorator that checks for a
// tx height timeout.
func NewTxTimeoutHeightDecorator() TxTimeoutHeightDecorator {
	return TxTimeoutHeightDecorator{}
}

// AnteHandle implements an AnteHandler decorator for the TxTimeoutHeightDecorator
// type where the current block height is checked against the tx's height timeout.
// If a height timeout is provided (non-zero) and is less than the current block
// height, then an error is returned.
func (txh TxTimeoutHeightDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	timeoutTx, ok := tx.(TxWithTimeoutHeight)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "expected tx to implement TxWithTimeoutHeight")
	}

	timeoutHeight := timeoutTx.GetTimeoutHeight()
	if timeoutHeight > 0 && uint64(ctx.BlockHeight()) > timeoutHeight {
		return ctx, errorsmod.Wrapf(
			sdkerrors.ErrTxTimeoutHeight, "block height: %d, timeout height: %d", ctx.BlockHeight(), timeoutHeight,
		)
	}

	timeoutTimestamp := timeoutTx.GetTimeoutTimeStamp()
	blockTime := ctx.BlockHeader().Time
	if !timeoutTimestamp.IsZero() && timeoutTimestamp.Unix() != 0 && timeoutTimestamp.Before(blockTime) {
		return ctx, errorsmod.Wrapf(
			sdkerrors.ErrTxTimeout, "block time: %s, timeout timestamp: %s", blockTime, timeoutTimestamp.String(),
		)
	}

	return next(ctx, tx, simulate)
}
