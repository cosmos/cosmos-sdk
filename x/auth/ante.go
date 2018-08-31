package auth

import (
	"bytes"
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
)

// NewAnteHandler returns an AnteHandler that checks
// and increments sequence numbers, checks signatures & account numbers,
// and deducts fees from the first signer.
// nolint: gocyclo
func NewAnteHandler(am AccountMapper, fck FeeCollectionKeeper) sdk.AnteHandler {

	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		// This AnteHandler requires Txs to be StdTxs
		stdTx, ok := tx.(StdTx)
		if !ok {
			return ctx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		// set the gas meter
		if simulate {
			newCtx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
		} else {
			newCtx = ctx.WithGasMeter(sdk.NewGasMeter(stdTx.Fee.Gas))
		}

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

		sigs := stdTx.GetSignatures() // When simulating, this would just be a 0-length slice.
		signerAddrs := stdTx.GetSigners()
		msgs := tx.GetMsgs()
		if simulate {
			sigs = make([]StdSignature, len(signerAddrs))
		}

		// charge gas for the memo
		newCtx.GasMeter().ConsumeGas(memoCostPerByte*sdk.Gas(len(stdTx.GetMemo())), "memo")

		// Get the sign bytes (requires all account & sequence numbers and the fee)
		sequences := make([]int64, len(sigs))
		accNums := make([]int64, len(sigs))
		for i := 0; i < len(sigs); i++ {
			sequences[i] = sigs[i].Sequence
			accNums[i] = sigs[i].AccountNumber
		}
		fee := stdTx.Fee

		// Check sig and nonce and collect signer accounts.
		var signerAccs = make([]Account, len(signerAddrs))
		for i := 0; i < len(sigs); i++ {
			signerAddr, sig := signerAddrs[i], sigs[i]

			// check signature, return account with incremented nonce
			signBytes := StdSignBytes(newCtx.ChainID(), accNums[i], sequences[i], fee, msgs, stdTx.GetMemo())
			signerAcc, res := processSig(newCtx, am, signerAddr, sig, signBytes, simulate)
			if !res.IsOK() {
				return newCtx, res, true
			}

			// first sig pays the fees
			// TODO: Add min fees
			// Can this function be moved outside of the loop?
			if i == 0 && !fee.Amount.IsZero() {
				newCtx.GasMeter().ConsumeGas(deductFeesCost, "deductFees")
				signerAcc, res = deductFees(signerAcc, fee)
				if !res.IsOK() {
					return newCtx, res, true
				}
				fck.addCollectedFees(newCtx, fee.Amount)
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

	// Perform checks that wouldn't pass successfully in simulation, i.e. sig
	// would be empty as simulated transactions come with no signatures whatsoever.
	if !simulate {
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
	}
	// Increment sequence number
	err := acc.SetSequence(seq + 1)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	// If pubkey is not known for account,
	// set it from the StdSignature.
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
			return secp256k1.GenPrivKey().PubKey(), sdk.Result{}
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

// BurnFeeHandler burns all fees (decreasing total supply)
func BurnFeeHandler(_ sdk.Context, _ sdk.Tx, _ sdk.Coins) {}
