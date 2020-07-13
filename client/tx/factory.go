package tx

import (
	"io"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// Factory defines a client transaction factory that facilitates generating and
// signing an application-specific transaction.
type Factory struct {
	keybase            keyring.Keyring
	txGenerator        client.TxGenerator
	accountRetriever   client.AccountRetriever
	accountNumber      uint64
	sequence           uint64
	gas                uint64
	gasAdjustment      float64
	chainID            string
	memo               string
	fees               sdk.Coins
	gasPrices          sdk.DecCoins
	signMode           signing.SignMode
	simulateAndExecute bool
}

const (
	signModeDirect    = "direct"
	signModeAminoJSON = "amino-json"
)

func NewFactoryCLI(clientCtx client.Context, flagSet *pflag.FlagSet) Factory {
	signModeStr, _ := flagSet.GetString(flags.FlagSignMode)

	signMode := signing.SignMode_SIGN_MODE_UNSPECIFIED
	switch signModeStr {
	case signModeDirect:
		signMode = signing.SignMode_SIGN_MODE_DIRECT
	case signModeAminoJSON:
		signMode = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	}

	accNum, _ := flagSet.GetUint64(flags.FlagAccountNumber)
	accSeq, _ := flagSet.GetUint64(flags.FlagSequence)
	gasAdj, _ := flagSet.GetFloat64(flags.FlagGasAdjustment)
	memo, _ := flagSet.GetString(flags.FlagMemo)

	gasStr, _ := flagSet.GetString(flags.FlagGas)
	gasSetting, _ := flags.ParseGasSetting(gasStr)

	f := Factory{
		txGenerator:        clientCtx.TxGenerator,
		accountRetriever:   clientCtx.AccountRetriever,
		keybase:            clientCtx.Keyring,
		chainID:            clientCtx.ChainID,
		gas:                gasSetting.Gas,
		simulateAndExecute: gasSetting.Simulate,
		accountNumber:      accNum,
		sequence:           accSeq,
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

// TODO: Remove in favor of NewFactoryCLI
func NewFactoryFromDeprecated(input io.Reader) Factory {
	kb, err := keyring.New(
		sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend),
		viper.GetString(flags.FlagHome),
		input,
	)
	if err != nil {
		panic(err)
	}

	signModeStr := viper.GetString(flags.FlagSignMode)
	signMode := signing.SignMode_SIGN_MODE_UNSPECIFIED
	switch signModeStr {
	case signModeDirect:
		signMode = signing.SignMode_SIGN_MODE_DIRECT
	case signModeAminoJSON:
		signMode = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	}

	gasSetting, _ := flags.ParseGasSetting(viper.GetString(flags.FlagGas))

	f := Factory{
		keybase:            kb,
		chainID:            viper.GetString(flags.FlagChainID),
		accountNumber:      viper.GetUint64(flags.FlagAccountNumber),
		sequence:           viper.GetUint64(flags.FlagSequence),
		gas:                gasSetting.Gas,
		simulateAndExecute: gasSetting.Simulate,
		gasAdjustment:      viper.GetFloat64(flags.FlagGasAdjustment),
		memo:               viper.GetString(flags.FlagMemo),
		signMode:           signMode,
	}

	f = f.WithFees(viper.GetString(flags.FlagFees))
	f = f.WithGasPrices(viper.GetString(flags.FlagGasPrices))

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

// SimulateAndExecute returns the option to simulate and then execute the transaction
// using the gas from the simulation results
func (f Factory) SimulateAndExecute() bool { return f.simulateAndExecute }

// WithTxGenerator returns a copy of the Factory with an updated TxGenerator.
func (f Factory) WithTxGenerator(g client.TxGenerator) Factory {
	f.txGenerator = g
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
	parsedFees, err := sdk.ParseCoins(fees)
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
