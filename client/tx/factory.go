package tx

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// Factory defines a client transaction factory that facilitates generating and
// signing an application-specific transaction.
type Factory struct {
	keybase            keyring.Keyring
	txConfig           client.TxConfig
	accountRetriever   client.AccountRetriever
	accountNumber      uint64
	sequence           uint64
	gas                uint64
	timeoutHeight      uint64
	gasAdjustment      float64
	chainID            string
	memo               string
	fees               sdk.Coins
	gasPrices          sdk.DecCoins
	signMode           signing.SignMode
	simulateAndExecute bool
}

// NewFactoryCLI creates a new Factory.
func NewFactoryCLI(clientCtx client.Context, flagSet *pflag.FlagSet) Factory {
	signModeStr := clientCtx.SignModeStr

	signMode := signing.SignMode_SIGN_MODE_UNSPECIFIED
	switch signModeStr {
	case flags.SignModeDirect:
		signMode = signing.SignMode_SIGN_MODE_DIRECT
	case flags.SignModeLegacyAminoJSON:
		signMode = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	}

	accNum, _ := flagSet.GetUint64(flags.FlagAccountNumber)
	accSeq, _ := flagSet.GetUint64(flags.FlagSequence)
	gasAdj, _ := flagSet.GetFloat64(flags.FlagGasAdjustment)
	memo, _ := flagSet.GetString(flags.FlagNote)
	timeoutHeight, _ := flagSet.GetUint64(flags.FlagTimeoutHeight)

	gasStr, _ := flagSet.GetString(flags.FlagGas)
	gasSetting, _ := flags.ParseGasSetting(gasStr)

	f := Factory{
		txConfig:           clientCtx.TxConfig,
		accountRetriever:   clientCtx.AccountRetriever,
		keybase:            clientCtx.Keyring,
		chainID:            clientCtx.ChainID,
		gas:                gasSetting.Gas,
		simulateAndExecute: gasSetting.Simulate,
		accountNumber:      accNum,
		sequence:           accSeq,
		timeoutHeight:      timeoutHeight,
		gasAdjustment:      gasAdj,
		memo:               memo,
		signMode:           signMode,
	}

	feesStr, _ := flagSet.GetString(flags.FlagFees)
	f = f.WithFees(feesStr)

	gasPricesStr, _ := flagSet.GetString(flags.FlagGasPrices)
	f = f.WithGasPrices(gasPricesStr)

	return f
}

func (f Factory) AccountNumber() uint64                     { return f.accountNumber }
func (f Factory) Sequence() uint64                          { return f.sequence }
func (f Factory) Gas() uint64                               { return f.gas }
func (f Factory) GasAdjustment() float64                    { return f.gasAdjustment }
func (f Factory) Keybase() keyring.Keyring                  { return f.keybase }
func (f Factory) ChainID() string                           { return f.chainID }
func (f Factory) Memo() string                              { return f.memo }
func (f Factory) Fees() sdk.Coins                           { return f.fees }
func (f Factory) GasPrices() sdk.DecCoins                   { return f.gasPrices }
func (f Factory) AccountRetriever() client.AccountRetriever { return f.accountRetriever }
func (f Factory) TimeoutHeight() uint64                     { return f.timeoutHeight }

// SimulateAndExecute returns the option to simulate and then execute the transaction
// using the gas from the simulation results
func (f Factory) SimulateAndExecute() bool { return f.simulateAndExecute }

// WithTxConfig returns a copy of the Factory with an updated TxConfig.
func (f Factory) WithTxConfig(g client.TxConfig) Factory {
	f.txConfig = g
	return f
}

// WithAccountRetriever returns a copy of the Factory with an updated AccountRetriever.
func (f Factory) WithAccountRetriever(ar client.AccountRetriever) Factory {
	f.accountRetriever = ar
	return f
}

// WithChainID returns a copy of the Factory with an updated chainID.
func (f Factory) WithChainID(chainID string) Factory {
	f.chainID = chainID
	return f
}

// WithGas returns a copy of the Factory with an updated gas value.
func (f Factory) WithGas(gas uint64) Factory {
	f.gas = gas
	return f
}

// WithFees returns a copy of the Factory with an updated fee.
func (f Factory) WithFees(fees string) Factory {
	parsedFees, err := sdk.ParseCoinsNormalized(fees)
	if err != nil {
		panic(err)
	}

	f.fees = parsedFees
	return f
}

// WithGasPrices returns a copy of the Factory with updated gas prices.
func (f Factory) WithGasPrices(gasPrices string) Factory {
	parsedGasPrices, err := sdk.ParseDecCoins(gasPrices)
	if err != nil {
		panic(err)
	}

	f.gasPrices = parsedGasPrices
	return f
}

// WithKeybase returns a copy of the Factory with updated Keybase.
func (f Factory) WithKeybase(keybase keyring.Keyring) Factory {
	f.keybase = keybase
	return f
}

// WithSequence returns a copy of the Factory with an updated sequence number.
func (f Factory) WithSequence(sequence uint64) Factory {
	f.sequence = sequence
	return f
}

// WithMemo returns a copy of the Factory with an updated memo.
func (f Factory) WithMemo(memo string) Factory {
	f.memo = memo
	return f
}

// WithAccountNumber returns a copy of the Factory with an updated account number.
func (f Factory) WithAccountNumber(accnum uint64) Factory {
	f.accountNumber = accnum
	return f
}

// WithGasAdjustment returns a copy of the Factory with an updated gas adjustment.
func (f Factory) WithGasAdjustment(gasAdj float64) Factory {
	f.gasAdjustment = gasAdj
	return f
}

// WithSimulateAndExecute returns a copy of the Factory with an updated gas
// simulation value.
func (f Factory) WithSimulateAndExecute(sim bool) Factory {
	f.simulateAndExecute = sim
	return f
}

// SignMode returns the sign mode configured in the Factory
func (f Factory) SignMode() signing.SignMode {
	return f.signMode
}

// WithSignMode returns a copy of the Factory with an updated sign mode value.
func (f Factory) WithSignMode(mode signing.SignMode) Factory {
	f.signMode = mode
	return f
}

// WithTimeoutHeight returns a copy of the Factory with an updated timeout height.
func (f Factory) WithTimeoutHeight(height uint64) Factory {
	f.timeoutHeight = height
	return f
}

// BuildUnsignedTx builds a transaction to be signed given a set of messages.
// Once created, the fee, memo, and messages are set.
func (f Factory) BuildUnsignedTx(msgs ...sdk.Msg) (client.TxBuilder, error) {
	if f.chainID == "" {
		return nil, fmt.Errorf("chain ID required but not specified")
	}

	fees := f.fees

	if !f.gasPrices.IsZero() {
		if !fees.IsZero() {
			return nil, errors.New("cannot provide both fees and gas prices")
		}

		glDec := sdk.NewDec(int64(f.gas))

		// Derive the fees based on the provided gas prices, where
		// fee = ceil(gasPrice * gasLimit).
		fees = make(sdk.Coins, len(f.gasPrices))

		for i, gp := range f.gasPrices {
			fee := gp.Amount.Mul(glDec)
			fees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	tx := f.txConfig.NewTxBuilder()

	if err := tx.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	tx.SetMemo(f.memo)
	tx.SetFeeAmount(fees)
	tx.SetGasLimit(f.gas)
	tx.SetTimeoutHeight(f.TimeoutHeight())

	return tx, nil
}

// PrintUnsignedTx will generate an unsigned transaction and print it to the writer
// specified by ctx.Output. If simulation was requested, the gas will be
// simulated and also printed to the same writer before the transaction is
// printed.
func (f Factory) PrintUnsignedTx(clientCtx client.Context, msgs ...sdk.Msg) error {
	if f.SimulateAndExecute() {
		if clientCtx.Offline {
			return errors.New("cannot estimate gas in offline mode")
		}

		_, adjusted, err := CalculateGas(clientCtx, f, msgs...)
		if err != nil {
			return err
		}

		f = f.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: f.Gas()})
	}

	tx, err := f.BuildUnsignedTx(msgs...)
	if err != nil {
		return err
	}

	json, err := clientCtx.TxConfig.TxJSONEncoder()(tx.GetTx())
	if err != nil {
		return err
	}

	return clientCtx.PrintString(fmt.Sprintf("%s\n", json))
}

// BuildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be
// built.
func (f Factory) BuildSimTx(msgs ...sdk.Msg) ([]byte, error) {
	txb, err := f.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		PubKey: &secp256k1.PubKey{},
		Data: &signing.SingleSignatureData{
			SignMode: f.signMode,
		},
		Sequence: f.Sequence(),
	}
	if err := txb.SetSignatures(sig); err != nil {
		return nil, err
	}

	return f.txConfig.TxEncoder()(txb.GetTx())
}

// Prepare ensures the account defined by ctx.GetFromAddress() exists and
// if the account number and/or the account sequence number are zero (not set),
// they will be queried for and set on the provided Factory. A new Factory with
// the updated fields will be returned.
func (f Factory) Prepare(clientCtx client.Context) (Factory, error) {
	fc := f

	from := clientCtx.GetFromAddress()

	if err := fc.accountRetriever.EnsureExists(clientCtx, from); err != nil {
		return fc, err
	}

	initNum, initSeq := fc.accountNumber, fc.sequence
	if initNum == 0 || initSeq == 0 {
		num, seq, err := fc.accountRetriever.GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return fc, err
		}

		if initNum == 0 {
			fc = fc.WithAccountNumber(num)
		}

		if initSeq == 0 {
			fc = fc.WithSequence(seq)
		}
	}

	return fc, nil
}
