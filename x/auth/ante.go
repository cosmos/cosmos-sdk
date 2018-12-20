package auth

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// TODO: Allow this to be configurable in the same way as minimum fees.
	// ref: https://github.com/cosmos/cosmos-sdk/issues/3101
	gasPerUnitCost uint64 = 10000 // how much gas = 1 atom
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(ak AccountKeeper, fck FeeCollectionKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		// all transactions must be of type auth.StdTx
		stdTx, ok := tx.(StdTx)
		if !ok {
			return ctx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		params := ak.GetParams(ctx)

		// Ensure that the provided fees meet a minimum threshold for the validator,
		// if this is a CheckTx. This is only for local mempool purposes, and thus
		// is only ran on check tx.
		if ctx.IsCheckTx() && !simulate {
			res := EnsureSufficientMempoolFees(ctx, stdTx)
			if !res.IsOK() {
				return newCtx, res, true
			}
		}

		newCtx = SetGasMeter(simulate, ctx, stdTx)

		// AnteHandlers must have their own defer/recover in order for the BaseApp
		// to know how much gas was used! This is because the GasMeter is created in
		// the AnteHandler, but if it panics the context won't be set properly in
		// runTx's recover call.
		defer func() {
			if r := recover(); r != nil {
				switch rType := r.(type) {
				case sdk.ErrorOutOfGas:
					log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
					res = sdk.ErrOutOfGas(log).Result()
					res.GasWanted = stdTx.Fee.Gas
					res.GasUsed = newCtx.GasMeter().GasConsumed()
					abort = true
				default:
					panic(r)
				}
			}
		}()

		if err := tx.ValidateBasic(); err != nil {
			return newCtx, err.Result(), true
		}

		if res := ValidateMemo(newCtx.GasMeter(), stdTx, params); !res.IsOK() {
			return newCtx, res, true
		}

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		signerAddrs := stdTx.GetSigners()
		signerAccs := make([]Account, len(signerAddrs))
		isGenesis := ctx.BlockHeight() == 0

		// fetch first signer, who's going to pay the fees
		signerAccs[0], res = GetSignerAcc(newCtx, ak, signerAddrs[0])
		if !res.IsOK() {
			return newCtx, res, true
		}
		if !stdTx.Fee.Amount.IsZero() {
			signerAccs[0], res = DeductFees(signerAccs[0], stdTx.Fee)
			if !res.IsOK() {
				return newCtx, res, true
			}

			fck.AddCollectedFees(newCtx, stdTx.Fee.Amount)
		}

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		stdSigs := stdTx.GetSignatures()

		for i := 0; i < len(stdSigs); i++ {
			// skip the fee payer, account is cached and fees were deducted already
			if i != 0 {
				signerAccs[i], res = GetSignerAcc(newCtx, ak, signerAddrs[i])
				if !res.IsOK() {
					return newCtx, res, true
				}
			}

			// check signature, return account with incremented nonce
			signBytes := GetSignBytes(newCtx.ChainID(), stdTx, signerAccs[i], isGenesis)
			signerAccs[i], res = processSig(newCtx, signerAccs[i], stdSigs[i], signBytes, simulate, params)
			if !res.IsOK() {
				return newCtx, res, true
			}

			ak.SetAccount(newCtx, signerAccs[i])
		}

		// TODO: tx tags (?)
		return newCtx, sdk.Result{GasWanted: stdTx.Fee.Gas}, false // continue...
	}
}

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak AccountKeeper, addr sdk.AccAddress) (Account, sdk.Result) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, sdk.Result{}
	}
	return nil, sdk.ErrUnknownAddress(addr.String()).Result()
}

// ValidateMemo validates the memo and if successful consumes gas for
// verification.
func ValidateMemo(gasMeter sdk.GasMeter, stdTx StdTx, params Params) sdk.Result {
	memoLength := len(stdTx.GetMemo())
	if uint64(memoLength) > params.MaxMemoCharacters {
		return sdk.ErrMemoTooLarge(
			fmt.Sprintf(
				"maximum number of characters is %d but received %d characters",
				params.MaxMemoCharacters, memoLength,
			),
		).Result()
	}

	gasMeter.ConsumeGas(params.MemoCostPerByte*sdk.Gas(memoLength), "memo")
	return sdk.Result{}
}

// verify the signature and increment the sequence. If the account doesn't have
// a pubkey, set it.
func processSig(
	ctx sdk.Context, acc Account, sig StdSignature, signBytes []byte, simulate bool, params Params,
) (updatedAcc Account, res sdk.Result) {

	pubKey, res := ProcessPubKey(acc, sig, simulate)
	if !res.IsOK() {
		return nil, res
	}

	err := acc.SetPubKey(pubKey)
	if err != nil {
		return nil, sdk.ErrInternal("setting PubKey on signer's account").Result()
	}

	consumeSignatureVerificationGas(ctx.GasMeter(), pubKey, params)
	if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("signature verification failed").Result()
	}

	err = acc.SetSequence(acc.GetSequence() + 1)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}

	return acc, res
}

var dummySecp256k1Pubkey secp256k1.PubKeySecp256k1

func init() {
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(dummySecp256k1Pubkey[:], bz)
}

// ProcessPubKey verifies that the given account address matches that of the
// StdSignature. In addition, it will set the public key of the account if it
// has not been set.
func ProcessPubKey(acc Account, sig StdSignature, simulate bool) (crypto.PubKey, sdk.Result) {
	// If pubkey is not known for account, set it from the StdSignature.
	pubKey := acc.GetPubKey()
	if simulate {
		// In simulate mode the transaction comes with no signatures, thus if the
		// account's pubkey is nil, both signature verification and gasKVStore.Set()
		// shall consume the largest amount, i.e. it takes more gas to verify
		// secp256k1 keys than ed25519 ones.
		if pubKey == nil {
			return dummySecp256k1Pubkey, sdk.Result{}
		}

		return pubKey, sdk.Result{}
	}

	if pubKey == nil {
		pubKey = sig.PubKey
		if pubKey == nil {
			return nil, sdk.ErrInvalidPubKey("PubKey not found").Result()
		}

		if !bytes.Equal(pubKey.Address(), acc.GetAddress()) {
			return nil, sdk.ErrInvalidPubKey(
				fmt.Sprintf("PubKey does not match Signer address %v", acc.GetAddress())).Result()
		}
	}

	return pubKey, sdk.Result{}
}

// consumeSignatureVerificationGas consumes gas for signature verification based
// upon the public key type. The cost is fetched from the given params and is
// matched by the concrete type.
//
// TODO: Design a cleaner and flexible way to match concrete public key types.
func consumeSignatureVerificationGas(meter sdk.GasMeter, pubkey crypto.PubKey, params Params) {
	pubkeyType := strings.ToLower(fmt.Sprintf("%T", pubkey))
	switch {
	case strings.Contains(pubkeyType, "ed25519"):
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
	case strings.Contains(pubkeyType, "secp256k1"):
		meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: secp256k1")
	default:
		panic(fmt.Sprintf("unrecognized signature type: %s", pubkeyType))
	}
}

func adjustFeesByGas(fees sdk.Coins, gas uint64) sdk.Coins {
	gasCost := gas / gasPerUnitCost
	gasFees := make(sdk.Coins, len(fees))

	// TODO: Make this not price all coins in the same way
	// TODO: Undo int64 casting once unsigned integers are supported for coins
	for i := 0; i < len(fees); i++ {
		gasFees[i] = sdk.NewInt64Coin(fees[i].Denom, int64(gasCost))
	}

	return fees.Plus(gasFees)
}

// DeductFees deducts fees from the given account.
//
// NOTE: We could use the CoinKeeper (in addition to the AccountKeeper, because
// the CoinKeeper doesn't give us accounts), but it seems easier to do this.
func DeductFees(acc Account, fee StdFee) (Account, sdk.Result) {
	coins := acc.GetCoins()
	feeAmount := fee.Amount

	if !feeAmount.IsValid() {
		return nil, sdk.ErrInsufficientFee(fmt.Sprintf("invalid fee amount: %s", feeAmount)).Result()
	}

	newCoins, ok := coins.SafeMinus(feeAmount)
	if ok {
		errMsg := fmt.Sprintf("%s < %s", coins, feeAmount)
		return nil, sdk.ErrInsufficientFunds(errMsg).Result()
	}

	err := acc.SetCoins(newCoins)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}

	return acc, sdk.Result{}
}

// EnsureSufficientMempoolFees verifies that the given transaction has supplied
// enough fees to cover a proposer's minimum fees. An result object is returned
// indicating success or failure.
//
// NOTE: This should only be called during CheckTx as it cannot be part of
// consensus.
func EnsureSufficientMempoolFees(ctx sdk.Context, stdTx StdTx) sdk.Result {
	// Currently we use a very primitive gas pricing model with a constant
	// gasPrice where adjustFeesByGas handles calculating the amount of fees
	// required based on the provided gas.
	//
	// TODO:
	// - Make the gasPrice not a constant, and account for tx size.
	// - Make Gas an unsigned integer and use tx basic validation
	if stdTx.Fee.Gas <= 0 {
		return sdk.ErrInternal(fmt.Sprintf("gas supplied must be a positive integer: %d", stdTx.Fee.Gas)).Result()
	}
	requiredFees := adjustFeesByGas(ctx.MinimumFees(), stdTx.Fee.Gas)

	// NOTE: !A.IsAllGTE(B) is not the same as A.IsAllLT(B).
	if !ctx.MinimumFees().IsZero() && !stdTx.Fee.Amount.IsAllGTE(requiredFees) {
		// validators reject any tx from the mempool with less than the minimum fee per gas * gas factor
		return sdk.ErrInsufficientFee(
			fmt.Sprintf(
				"insufficient fee, got: %q required: %q", stdTx.Fee.Amount, requiredFees),
		).Result()
	}

	return sdk.Result{}
}

// SetGasMeter returns a new context with a gas meter set from a given context.
func SetGasMeter(simulate bool, ctx sdk.Context, stdTx StdTx) sdk.Context {
	// In various cases such as simulation and during the genesis block, we do not
	// meter any gas utilization.
	if simulate || ctx.BlockHeight() == 0 {
		return ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	}

	return ctx.WithGasMeter(sdk.NewGasMeter(stdTx.Fee.Gas))
}

// GetSignBytes returns a slice of bytes to sign over for a given transaction
// and an account.
func GetSignBytes(chainID string, stdTx StdTx, acc Account, genesis bool) []byte {
	var accNum uint64
	if !genesis {
		accNum = acc.GetAccountNumber()
	}

	return StdSignBytes(
		chainID, accNum, acc.GetSequence(), stdTx.Fee, stdTx.Msgs, stdTx.Memo,
	)
}
