package types

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxBuilder implements a transaction context created in SDK modules.
type TxBuilder struct {
	txEncoder          sdk.TxEncoder
	keybase            keyring.Keyring
	accountNumber      uint64
	sequence           uint64
	gas                uint64
	gasAdjustment      float64
	simulateAndExecute bool
	chainID            string
	memo               string
	fees               sdk.Coins
	gasPrices          sdk.DecCoins
}

// NewTxBuilder returns a new initialized TxBuilder.
func NewTxBuilder(
	txEncoder sdk.TxEncoder, accNumber, seq, gas uint64, gasAdj float64,
	simulateAndExecute bool, chainID, memo string, fees sdk.Coins, gasPrices sdk.DecCoins,
) TxBuilder {

	return TxBuilder{
		txEncoder:          txEncoder,
		keybase:            nil,
		accountNumber:      accNumber,
		sequence:           seq,
		gas:                gas,
		gasAdjustment:      gasAdj,
		simulateAndExecute: simulateAndExecute,
		chainID:            chainID,
		memo:               memo,
		fees:               fees,
		gasPrices:          gasPrices,
	}
}

// NewTxBuilderFromCLI returns a new initialized TxBuilder with parameters from
// the command line using Viper.
func NewTxBuilderFromCLI(input io.Reader) TxBuilder {
	kb, err := keyring.New(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), input)
	if err != nil {
		panic(err)
	}
	txbldr := TxBuilder{
		keybase:            kb,
		accountNumber:      viper.GetUint64(flags.FlagAccountNumber),
		sequence:           viper.GetUint64(flags.FlagSequence),
		gas:                flags.GasFlagVar.Gas,
		gasAdjustment:      viper.GetFloat64(flags.FlagGasAdjustment),
		simulateAndExecute: flags.GasFlagVar.Simulate,
		chainID:            viper.GetString(flags.FlagChainID),
		memo:               viper.GetString(flags.FlagMemo),
	}

	txbldr = txbldr.WithFees(viper.GetString(flags.FlagFees))
	txbldr = txbldr.WithGasPrices(viper.GetString(flags.FlagGasPrices))

	return txbldr
}

// TxEncoder returns the transaction encoder
func (bldr TxBuilder) TxEncoder() sdk.TxEncoder { return bldr.txEncoder }

// AccountNumber returns the account number
func (bldr TxBuilder) AccountNumber() uint64 { return bldr.accountNumber }

// Sequence returns the transaction sequence
func (bldr TxBuilder) Sequence() uint64 { return bldr.sequence }

// Gas returns the gas for the transaction
func (bldr TxBuilder) Gas() uint64 { return bldr.gas }

// GasAdjustment returns the gas adjustment
func (bldr TxBuilder) GasAdjustment() float64 { return bldr.gasAdjustment }

// Keybase returns the keybase
func (bldr TxBuilder) Keybase() keyring.Keyring { return bldr.keybase }

// SimulateAndExecute returns the option to simulate and then execute the transaction
// using the gas from the simulation results
func (bldr TxBuilder) SimulateAndExecute() bool { return bldr.simulateAndExecute }

// ChainID returns the chain id
func (bldr TxBuilder) ChainID() string { return bldr.chainID }

// Memo returns the memo message
func (bldr TxBuilder) Memo() string { return bldr.memo }

// Fees returns the fees for the transaction
func (bldr TxBuilder) Fees() sdk.Coins { return bldr.fees }

// GasPrices returns the gas prices set for the transaction, if any.
func (bldr TxBuilder) GasPrices() sdk.DecCoins { return bldr.gasPrices }

// WithTxEncoder returns a copy of the context with an updated codec.
func (bldr TxBuilder) WithTxEncoder(txEncoder sdk.TxEncoder) TxBuilder {
	bldr.txEncoder = txEncoder
	return bldr
}

// WithChainID returns a copy of the context with an updated chainID.
func (bldr TxBuilder) WithChainID(chainID string) TxBuilder {
	bldr.chainID = chainID
	return bldr
}

// WithGas returns a copy of the context with an updated gas.
func (bldr TxBuilder) WithGas(gas uint64) TxBuilder {
	bldr.gas = gas
	return bldr
}

// WithFees returns a copy of the context with an updated fee.
func (bldr TxBuilder) WithFees(fees string) TxBuilder {
	parsedFees, err := sdk.ParseCoins(fees)
	if err != nil {
		panic(err)
	}

	bldr.fees = parsedFees
	return bldr
}

// WithGasPrices returns a copy of the context with updated gas prices.
func (bldr TxBuilder) WithGasPrices(gasPrices string) TxBuilder {
	parsedGasPrices, err := sdk.ParseDecCoins(gasPrices)
	if err != nil {
		panic(err)
	}

	bldr.gasPrices = parsedGasPrices
	return bldr
}

// WithKeybase returns a copy of the context with updated keybase.
func (bldr TxBuilder) WithKeybase(keybase keyring.Keyring) TxBuilder {
	bldr.keybase = keybase
	return bldr
}

// WithSequence returns a copy of the context with an updated sequence number.
func (bldr TxBuilder) WithSequence(sequence uint64) TxBuilder {
	bldr.sequence = sequence
	return bldr
}

// WithMemo returns a copy of the context with an updated memo.
func (bldr TxBuilder) WithMemo(memo string) TxBuilder {
	bldr.memo = strings.TrimSpace(memo)
	return bldr
}

// WithAccountNumber returns a copy of the context with an account number.
func (bldr TxBuilder) WithAccountNumber(accnum uint64) TxBuilder {
	bldr.accountNumber = accnum
	return bldr
}

// BuildSignMsg builds a single message to be signed from a TxBuilder given a
// set of messages. It returns an error if a fee is supplied but cannot be
// parsed.
func (bldr TxBuilder) BuildSignMsg(msgs []sdk.Msg) (StdSignMsg, error) {
	if bldr.chainID == "" {
		return StdSignMsg{}, fmt.Errorf("chain ID required but not specified")
	}

	fees := bldr.fees
	if !bldr.gasPrices.IsZero() {
		if !fees.IsZero() {
			return StdSignMsg{}, errors.New("cannot provide both fees and gas prices")
		}

		glDec := sdk.NewDec(int64(bldr.gas))

		// Derive the fees based on the provided gas prices, where
		// fee = ceil(gasPrice * gasLimit).
		fees = make(sdk.Coins, len(bldr.gasPrices))
		for i, gp := range bldr.gasPrices {
			fee := gp.Amount.Mul(glDec)
			fees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	return StdSignMsg{
		ChainID:       bldr.chainID,
		AccountNumber: bldr.accountNumber,
		Sequence:      bldr.sequence,
		Memo:          bldr.memo,
		Msgs:          msgs,
		Fee:           NewStdFee(bldr.gas, fees),
	}, nil
}

// Sign signs a transaction given a name and a single message to be signed.
// Returns error if signing fails.
func (bldr TxBuilder) Sign(name string, msg StdSignMsg) ([]byte, error) {
	sig, err := MakeSignature(bldr.keybase, name, msg)
	if err != nil {
		return nil, err
	}

	return bldr.txEncoder(NewStdTx(msg.Msgs, msg.Fee, []StdSignature{sig}, msg.Memo))
}

// BuildAndSign builds a single message to be signed, and signs a transaction
// with the built message given a name and a set of messages.
func (bldr TxBuilder) BuildAndSign(name string, msgs []sdk.Msg) ([]byte, error) {
	msg, err := bldr.BuildSignMsg(msgs)
	if err != nil {
		return nil, err
	}

	return bldr.Sign(name, msg)
}

// BuildTxForSim creates a StdSignMsg and encodes a transaction with the
// StdSignMsg with a single empty StdSignature for tx simulation.
func (bldr TxBuilder) BuildTxForSim(msgs []sdk.Msg) ([]byte, error) {
	signMsg, err := bldr.BuildSignMsg(msgs)
	if err != nil {
		return nil, err
	}

	// the ante handler will populate with a sentinel pubkey
	sigs := []StdSignature{{}}
	return bldr.txEncoder(NewStdTx(signMsg.Msgs, signMsg.Fee, sigs, signMsg.Memo))
}

// SignStdTx appends a signature to a StdTx and returns a copy of it. If append
// is false, it replaces the signatures already attached with the new signature.
func (bldr TxBuilder) SignStdTx(name string, stdTx StdTx, appendSig bool) (signedStdTx StdTx, err error) {
	if bldr.chainID == "" {
		return StdTx{}, fmt.Errorf("chain ID required but not specified")
	}

	stdSignature, err := MakeSignature(bldr.keybase, name, StdSignMsg{
		ChainID:       bldr.chainID,
		AccountNumber: bldr.accountNumber,
		Sequence:      bldr.sequence,
		Fee:           stdTx.Fee,
		Msgs:          stdTx.GetMsgs(),
		Memo:          stdTx.GetMemo(),
	})
	if err != nil {
		return
	}

	sigs := stdTx.Signatures
	if len(sigs) == 0 || !appendSig {
		sigs = []StdSignature{stdSignature}
	} else {
		sigs = append(sigs, stdSignature)
	}
	signedStdTx = NewStdTx(stdTx.GetMsgs(), stdTx.Fee, sigs, stdTx.GetMemo())
	return
}

// MakeSignature builds a StdSignature given a keyring, key name, and a StdSignMsg.
func MakeSignature(kr keyring.Keyring, name string, msg StdSignMsg) (sig StdSignature, err error) {
	if kr == nil {
		kr, err = keyring.New(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), os.Stdin)
		if err != nil {
			return
		}
	}

	sigBytes, pubkey, err := kr.Sign(name, msg.Bytes())
	if err != nil {
		return
	}

	return StdSignature{
		PubKey:    pubkey.Bytes(),
		Signature: sigBytes,
	}, nil
}
