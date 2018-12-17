package auth

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	memoCostPerByte     sdk.Gas = 3
	ed25519VerifyCost           = 590
	secp256k1VerifyCost         = 1000
	maxMemoCharacters           = 256

	// how much gas = 1 atom
	gasPerUnitCost = 10000

	// max total number of sigs per tx
	txSigLimit = 7
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(am AccountKeeper, fck FeeCollectionKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		// all transactions must be of type auth.StdTx
		stdTx, ok := tx.(StdTx)
		if !ok {
			return ctx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

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

		newCtx.GasMeter().ConsumeGas(memoCostPerByte*sdk.Gas(len(stdTx.GetMemo())), "memo")

		signerAccs, res := GetSignerAccs(newCtx, am, stdTx.GetSigners())
		if !res.IsOK() {
			return newCtx, res, true
		}

		// the first signer pays the transaction fees
		if !stdTx.Fee.Amount.IsZero() {
			signerAccs[0], res = DeductFees(signerAccs[0], stdTx.Fee)
			if !res.IsOK() {
				return newCtx, res, true
			}

			fck.AddCollectedFees(newCtx, stdTx.Fee.Amount)
		}

		isGenesis := ctx.BlockHeight() == 0
		signBytesList := GetSignBytesList(newCtx.ChainID(), stdTx, signerAccs, isGenesis)

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		stdSigs := stdTx.GetSignatures()

		for i := 0; i < len(stdSigs); i++ {
			// check signature, return account with incremented nonce
			signerAccs[i], res = processSig(newCtx, signerAccs[i], stdSigs[i], signBytesList[i], simulate)
			if !res.IsOK() {
				return newCtx, res, true
			}

			am.SetAccount(newCtx, signerAccs[i])
		}

		// cache the signer accounts in the context
		newCtx = WithSigners(newCtx, signerAccs)

		// TODO: tx tags (?)
		return newCtx, sdk.Result{GasWanted: stdTx.Fee.Gas}, false // continue...
	}
}

// GetSignerAccs returns a list of signers for a given list of addresses that
// are expected to sign a transaction.
func GetSignerAccs(ctx sdk.Context, am AccountKeeper, addrs []sdk.AccAddress) (accs []Account, res sdk.Result) {
	accs = make([]Account, len(addrs))
	for i := 0; i < len(accs); i++ {
		accs[i] = am.GetAccount(ctx, addrs[i])
		if accs[i] == nil {
			return nil, sdk.ErrUnknownAddress(addrs[i].String()).Result()
		}
	}

	return
}

// verify the signature and increment the sequence. If the account doesn't have
// a pubkey, set it.
func processSig(
	ctx sdk.Context, acc Account, sig StdSignature, signBytes []byte, simulate bool,
) (updatedAcc Account, res sdk.Result) {

	pubKey, res := ProcessPubKey(acc, sig, simulate)
	if !res.IsOK() {
		return nil, res
	}

	err := acc.SetPubKey(pubKey)
	if err != nil {
		return nil, sdk.ErrInternal("setting PubKey on signer's account").Result()
	}

	consumeSignatureVerificationGas(ctx.GasMeter(), pubKey)
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

func consumeSignatureVerificationGas(meter sdk.GasMeter, pubkey crypto.PubKey) {
	switch pubkey.(type) {
	case ed25519.PubKeyEd25519:
		meter.ConsumeGas(ed25519VerifyCost, "ante verify: ed25519")
	case secp256k1.PubKeySecp256k1:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: secp256k1")
	default:
		panic("Unrecognized signature type")
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
		return sdk.ErrInternal(fmt.Sprintf("invalid gas supplied: %d", stdTx.Fee.Gas)).Result()
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

// GetSignBytesList returns a slice of bytes to sign over for a given transaction
// and list of accounts.
func GetSignBytesList(chainID string, stdTx StdTx, accs []Account, genesis bool) (signatureBytesList [][]byte) {
	signatureBytesList = make([][]byte, len(accs))
	for i := 0; i < len(accs); i++ {
		accNum := accs[i].GetAccountNumber()
		if genesis {
			accNum = 0
		}
		signatureBytesList[i] = StdSignBytes(chainID,
			accNum, accs[i].GetSequence(),
			stdTx.Fee, stdTx.Msgs, stdTx.Memo)
	}
	return
}
