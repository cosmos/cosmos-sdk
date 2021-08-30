package ante

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

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
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	params := cgts.ak.GetParams(ctx)

	ctx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*sdk.Gas(len(ctx.TxBytes())), "txSize")

	// simulate gas cost for signatures in simulate mode
	if simulate {
		// in simulate mode, each element should be a nil signature
		sigs, err := sigTx.GetSignaturesV2()
		if err != nil {
			return ctx, err
		}
		n := len(sigs)

		for i, signer := range sigTx.GetSigners() {
			// if signature is already filled in, no need to simulate gas cost
			if i < n && !isIncompleteSignature(sigs[i].Data) {
				continue
			}

			var pubkey cryptotypes.PubKey

			acc := cgts.ak.GetAccount(ctx, signer)

			// use placeholder simSecp256k1Pubkey if sig is nil
			if acc == nil || acc.GetPubKey() == nil {
				pubkey = simSecp256k1Pubkey
			} else {
				pubkey = acc.GetPubKey()
			}

			// use stdsignature to mock the size of a full signature
			simSig := legacytx.StdSignature{ //nolint:staticcheck // this will be removed when proto is ready
				Signature: simSecp256k1Sig[:],
				PubKey:    pubkey,
			}

			sigBz := legacy.Cdc.MustMarshal(simSig)
			cost := sdk.Gas(len(sigBz) + 6)

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
		for _, s := range data.Signatures {
			if isIncompleteSignature(s) {
				return true
			}
		}
	}

	return false
}
