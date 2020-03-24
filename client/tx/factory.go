package tx

import (
	"io"
	"strings"

	"githuf.com/spf13/viper"
	"githuf.com/tendermint/tendermint/crypto"

	"githuf.com/cosmos/cosmos-sdk/client/flags"
	"githuf.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "githuf.com/cosmos/cosmos-sdk/types"
)

// AccountRetriever defines the interfaces required for use by the Factory to
// ensure an account exists and to be able to query for account fields necessary
// for signing.
type AccountRetriever interface {
	EnsureExists(addr sdk.AccAddress) error
	GetAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error)
}

// Factory defines a client transaction factory that facilitates generating and
// signing an application-specific transaction.
type Factory struct {
	keybase            keys.Keybase
	txGenerator        Generator
	accountRetriever   AccountRetriever
	feeFn              func(gas uint64, amount sdk.Coins) sdk.Fee
	sigFn              func(pk crypto.PubKey, sig []byte) sdk.Signature
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

func NewFactoryFromCLI(input io.Reader) Factory {
	kb, err := keys.NewKeyring(
		sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend),
		viper.GetString(flags.FlagHome),
		input,
	)
	if err != nil {
		panic(err)
	}

	f := Factory{
		keybase:            kb,
		accountNumber:      viper.GetUint64(flags.FlagAccountNumber),
		sequence:           viper.GetUint64(flags.FlagSequence),
		gas:                flags.GasFlagVar.Gas,
		gasAdjustment:      viper.GetFloat64(flags.FlagGasAdjustment),
		simulateAndExecute: flags.GasFlagVar.Simulate,
		chainID:            viper.GetString(flags.FlagChainID),
		memo:               viper.GetString(flags.FlagMemo),
	}

	f = f.WithFees(viper.GetString(flags.FlagFees))
	f = f.WithGasPrices(viper.GetString(flags.FlagGasPrices))

	return f
}

// nolint
func (f Factory) AccountNumber() uint64              { return f.accountNumber }
func (f Factory) Sequence() uint64                   { return f.sequence }
func (f Factory) Gas() uint64                        { return f.gas }
func (f Factory) GasAdjustment() float64             { return f.gasAdjustment }
func (f Factory) Keybase() keys.Keybase              { return f.keybase }
func (f Factory) ChainID() string                    { return f.chainID }
func (f Factory) Memo() string                       { return f.memo }
func (f Factory) Fees() sdk.Coins                    { return f.fees }
func (f Factory) GasPrices() sdk.DecCoins            { return f.gasPrices }
func (f Factory) AccountRetriever() AccountRetriever { return f.accountRetriever }

// SimulateAndExecute returns the option to simulate and then execute the transaction
// using the gas from the simulation results
func (f Factory) SimulateAndExecute() bool { return f.simulateAndExecute }

// WithTxGenerator returns a copy of the Builder with an updated Generator.
func (f Factory) WithTxGenerator(g Generator) Factory {
	f.txGenerator = g
	return f
}

// WithAccountRetriever returns a copy of the Builder with an updated AccountRetriever.
func (f Factory) WithAccountRetriever(ar AccountRetriever) Factory {
	f.accountRetriever = ar
	return f
}

// WithChainID returns a copy of the Builder with an updated chainID.
func (f Factory) WithChainID(chainID string) Factory {
	f.chainID = chainID
	return f
}

// WithGas returns a copy of the Builder with an updated gas value.
func (f Factory) WithGas(gas uint64) Factory {
	f.gas = gas
	return f
}

// WithSigFn returns a copy of the Builder with an updated tx signature constructor.
func (f Factory) WithSigFn(sigFn func(pk crypto.PubKey, sig []byte) sdk.Signature) Factory {
	f.sigFn = sigFn
	return f
}

// WithFeeFn returns a copy of the Builder with an updated fee constructor.
func (f Factory) WithFeeFn(feeFn func(gas uint64, amount sdk.Coins) sdk.Fee) Factory {
	f.feeFn = feeFn
	return f
}

// WithFees returns a copy of the Builder with an updated fee.
func (f Factory) WithFees(fees string) Factory {
	parsedFees, err := sdk.ParseCoins(fees)
	if err != nil {
		panic(err)
	}

	f.fees = parsedFees
	return f
}

// WithGasPrices returns a copy of the Builder with updated gas prices.
func (f Factory) WithGasPrices(gasPrices string) Factory {
	parsedGasPrices, err := sdk.ParseDecCoins(gasPrices)
	if err != nil {
		panic(err)
	}

	f.gasPrices = parsedGasPrices
	return f
}

// WithKeybase returns a copy of the Builder with updated Keybase.
func (f Factory) WithKeybase(keybase keys.Keybase) Factory {
	f.keybase = keybase
	return f
}

// WithSequence returns a copy of the Builder with an updated sequence number.
func (f Factory) WithSequence(sequence uint64) Factory {
	f.sequence = sequence
	return f
}

// WithMemo returns a copy of the Builder with an updated memo.
func (f Factory) WithMemo(memo string) Factory {
	f.memo = strings.TrimSpace(memo)
	return f
}

// WithAccountNumber returns a copy of the Builder with an updated account number.
func (f Factory) WithAccountNumber(accnum uint64) Factory {
	f.accountNumber = accnum
	return f
}
