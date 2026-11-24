package ante

import (
	"slices"
	"time"

	cmtmldsa65 "github.com/cometbft/cometbft/crypto/mldsa65"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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

			// prefer the pubkey stored on the account; for an account's first tx
			// fall back to the pubkey carried in the tx's SignerInfo, and only then
			// to the secp256k1 placeholder.
			switch {
			case acc != nil && acc.GetPubKey() != nil:
				pubkey = acc.GetPubKey()
			case i < n && sigs[i].PubKey != nil:
				pubkey = sigs[i].PubKey
			default:
				pubkey = simSecp256k1Pubkey
			}

			// use stdsignature to mock the size of a full signature
			simSig := legacytx.StdSignature{ //nolint:staticcheck // SA1019: legacytx.StdSignature is deprecated
				Signature: simSignaturePlaceholder(pubkey),
				PubKey:    pubkey,
			}

			sigBz, err := legacy.Cdc.Marshal(simSig)
			if err != nil {
				// pubkey's concrete type isn't registered with the legacy
				// Amino codec (e.g. secp256r1); size a registered placeholder
				// of the same byte length instead, so the result still
				// reflects the real key's size.
				sigBz = legacy.Cdc.MustMarshal(legacytx.StdSignature{ //nolint:staticcheck // SA1019: legacytx.StdSignature is deprecated
					Signature: simSig.Signature,
					PubKey:    &secp256k1.PubKey{Key: make([]byte, len(pubkey.Bytes()))},
				})
			}
			cost := storetypes.Gas(len(sigBz) + 6)

			// If the pubkey is a multi-signature pubkey, then we estimate for the maximum
			// number of signers.
			if _, ok := pubkey.(*multisig.LegacyAminoPubKey); ok {
				cost *= params.TxSigLimit
			}

			ctx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*cost, "txSize")
		}
	}

	return next(ctx, tx, simulate)
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

// simSignaturePlaceholder returns a zero signature of the size the given key
// type produces, so simulation charges tx-size gas for the real signature.
// Key types with signatures larger than secp256k1's must be listed here.
func simSignaturePlaceholder(pk cryptotypes.PubKey) []byte {
	if _, ok := pk.(*mldsa65.PubKey); ok {
		return make([]byte, cmtmldsa65.SignatureSize)
	}
	return simSecp256k1Sig[:]
}
