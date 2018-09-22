package auth

import (
	"bytes"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

const (
	deductFeesCost      sdk.Gas = 10
	memoCostPerByte     sdk.Gas = 1
	ed25519VerifyCost           = 59
	secp256k1VerifyCost         = 100
	maxMemoCharacters           = 100
	gasPrice                    = 0.001
)

// NewAnteHandler returns an AnteHandler that checks
// and increments sequence numbers, checks signatures & account numbers,
// and deducts fees from the first signer.
func NewAnteHandler(am AccountMapper, fck FeeCollectionKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		// This AnteHandler requires Txs to be StdTxs
		stdTx, ok := tx.(StdTx)
		if !ok {
			return ctx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		// Ensure that the provided fees meet a minimum threshold for the validator, if this is a CheckTx.
		// This is for mempool purposes, and thus is only ran on check tx.
		if ctx.IsCheckTx() && !simulate {
			res := ensureSufficientMempoolFees(ctx, stdTx)
			if !res.IsOK() {
				return newCtx, res, true
			}
		}

		newCtx = setGasMeter(simulate, ctx, stdTx)

		// AnteHandlers must have their own defer/recover in order
		// for the BaseApp to know how much gas was used!
		// This is because the GasMeter is created in the AnteHandler,
		// but if it panics the context won't be set properly in runTx's recover ...
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

		err := validateBasic(stdTx)
		if err != nil {
			return newCtx, err.Result(), true
		}
		// charge gas for the memo
		newCtx.GasMeter().ConsumeGas(memoCostPerByte*sdk.Gas(len(stdTx.GetMemo())), "memo")

		// stdSigs contains the sequence number, account number, and signatures
		stdSigs := stdTx.GetSignatures() // When simulating, this would just be a 0-length slice.
		signerAddrs := stdTx.GetSigners()

		// create the list of all sign bytes
		signBytesList := getSignBytesList(newCtx.ChainID(), stdTx, stdSigs)

		// Get the sign bytes (requires all account & sequence numbers and the fee)
		// Check sig and nonce and collect signer accounts.
		var signerAccs = make([]Account, len(signerAddrs))
		for i := 0; i < len(stdSigs); i++ {
			signerAddr, sig := signerAddrs[i], stdSigs[i]

			// check signature, return account with incremented nonce
			signerAcc, res := processSig(newCtx, am, signerAddr, sig, signBytesList[i], simulate)
			if !res.IsOK() {
				return newCtx, res, true
			}

			// first sig pays the fees
			// TODO: move outside of the loop
			if i == 0 && !stdTx.Fee.Amount.IsZero() {
				newCtx.GasMeter().ConsumeGas(deductFeesCost, "deductFees")
				signerAcc, res = deductFees(signerAcc, stdTx.Fee)
				if !res.IsOK() {
					return newCtx, res, true
				}
				fck.addCollectedFees(newCtx, stdTx.Fee.Amount)
			}

			// Save the account.
			am.SetAccount(newCtx, signerAcc)
			signerAccs[i] = signerAcc
		}

		// cache the signer accounts in the context
		newCtx = WithSigners(newCtx, signerAccs)

		// TODO: tx tags (?)

		return newCtx, sdk.Result{GasWanted: stdTx.Fee.Gas}, false // continue...
	}
}

// Validate the transaction based on things that don't depend on the context
func validateBasic(tx StdTx) (err sdk.Error) {
	// Assert that there are signatures.
	sigs := tx.GetSignatures()
	if len(sigs) == 0 {
		return sdk.ErrUnauthorized("no signers")
	}

	// Assert that number of signatures is correct.
	var signerAddrs = tx.GetSigners()
	if len(sigs) != len(signerAddrs) {
		return sdk.ErrUnauthorized("wrong number of signers")
	}

	memo := tx.GetMemo()
	if len(memo) > maxMemoCharacters {
		return sdk.ErrMemoTooLarge(
			fmt.Sprintf("maximum number of characters is %d but received %d characters",
				maxMemoCharacters, len(memo)))
	}
	return nil
}

// verify the signature and increment the sequence.
// if the account doesn't have a pubkey, set it.
// TODO: Change this function to already take in the account
func processSig(
	ctx sdk.Context, am AccountMapper,
	addr sdk.AccAddress, sig StdSignature, signBytes []byte, simulate bool) (
	acc Account, res sdk.Result) {
	// Get the account.
	acc = am.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(addr.String()).Result()
	}

	accnum := acc.GetAccountNumber()
	seq := acc.GetSequence()

	// Check account number.
	if accnum != sig.AccountNumber {
		return nil, sdk.ErrInvalidSequence(
			fmt.Sprintf("Invalid account number. Got %d, expected %d", sig.AccountNumber, accnum)).Result()
	}

	// Check sequence number.
	if seq != sig.Sequence {
		return nil, sdk.ErrInvalidSequence(
			fmt.Sprintf("Invalid sequence. Got %d, expected %d", sig.Sequence, seq)).Result()
	}
	err := acc.SetSequence(seq + 1)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	pubKey, res := processPubKey(acc, sig, simulate)
	if !res.IsOK() {
		return nil, res
	}
	err = acc.SetPubKey(pubKey)
	if err != nil {
		return nil, sdk.ErrInternal("setting PubKey on signer's account").Result()
	}

	consumeSignatureVerificationGas(ctx.GasMeter(), pubKey)
	if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("signature verification failed").Result()
	}

	return
}

var dummySecp256k1Pubkey secp256k1.PubKeySecp256k1

func init() {
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(dummySecp256k1Pubkey[:], bz)
}

func processPubKey(acc Account, sig StdSignature, simulate bool) (crypto.PubKey, sdk.Result) {
	// If pubkey is not known for account,
	// set it from the StdSignature.
	pubKey := acc.GetPubKey()
	if simulate {
		// In simulate mode the transaction comes with no signatures, thus
		// if the account's pubkey is nil, both signature verification
		// and gasKVStore.Set() shall consume the largest amount, i.e.
		// it takes more gas to verifiy secp256k1 keys than ed25519 ones.
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

func adjustFeesByGas(fees sdk.Coins, gas int64) sdk.Coins {
	gasCost := int64(float64(gas) * gasPrice)
	gasFees := make(sdk.Coins, len(fees))
	// TODO: Make this not price all coins in the same way
	for i := 0; i < len(fees); i++ {
		gasFees[i] = sdk.NewInt64Coin(fees[i].Denom, gasCost)
	}
	return fees.Plus(gasFees)
}

// Deduct the fee from the account.
// We could use the CoinKeeper (in addition to the AccountMapper,
// because the CoinKeeper doesn't give us accounts), but it seems easier to do this.
func deductFees(acc Account, fee StdFee) (Account, sdk.Result) {
	coins := acc.GetCoins()
	feeAmount := fee.Amount

	newCoins := coins.Minus(feeAmount)
	if !newCoins.IsNotNegative() {
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

func ensureSufficientMempoolFees(ctx sdk.Context, stdTx StdTx) sdk.Result {
	// currently we use a very primitive gas pricing model with a constant gasPrice.
	// adjustFeesByGas handles calculating the amount of fees required based on the provided gas.
	// TODO: Make the gasPrice not a constant, and account for tx size.
	requiredFees := adjustFeesByGas(ctx.MinimumFees(), stdTx.Fee.Gas)

	if !ctx.MinimumFees().IsZero() && stdTx.Fee.Amount.IsLT(requiredFees) {
		// validators reject any tx from the mempool with less than the minimum fee per gas * gas factor
		return sdk.ErrInsufficientFee(fmt.Sprintf(
			"insufficient fee, got: %q required: %q", stdTx.Fee.Amount, requiredFees)).Result()
	}
	return sdk.Result{}
}

func setGasMeter(simulate bool, ctx sdk.Context, stdTx StdTx) sdk.Context {
	// set the gas meter
	if simulate {
		return ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	}
	return ctx.WithGasMeter(sdk.NewGasMeter(stdTx.Fee.Gas))
}

func getSignBytesList(chainID string, stdTx StdTx, stdSigs []StdSignature) (
	signatureBytesList [][]byte) {
	signatureBytesList = make([][]byte, len(stdSigs))
	for i := 0; i < len(stdSigs); i++ {
		signatureBytesList[i] = StdSignBytes(chainID,
			stdSigs[i].AccountNumber, stdSigs[i].Sequence,
			stdTx.Fee, stdTx.Msgs, stdTx.Memo)
	}
	return
}

// BurnFeeHandler burns all fees (decreasing total supply)
func BurnFeeHandler(_ sdk.Context, _ sdk.Tx, _ sdk.Coins) {}
