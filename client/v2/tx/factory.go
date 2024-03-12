package tx

import (
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/math"
	authsigning "cosmossdk.io/x/auth/signing"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/go-bip39"
	"github.com/spf13/pflag"
	"math/big"
	"os"
	"strings"
)

// TxCodecProvider defines an interface that contains transaction
// encoders and decoders
type TxCodecProvider interface {
	TxEncoder() sdk.TxV2Encoder
	TxDecoder() sdk.TxV2Decoder
	TxJSONEncoder() sdk.TxV2Encoder
	TxJSONDecoder() sdk.TxV2Decoder
}

// Factory defines a client transaction factory that facilitates generating and
// signing an application-specific transaction.
type Factory struct {
	keybase           keyring.Keyring
	txBuilderProvider TxBuilderProvider
	codec             TxCodecProvider
	accountRetriever  AccountRetriever
	signingConfig     TxSigningConfig
	TxParameters
}

func NewFactoryCLI(clientCtx txContext, flagSet *pflag.FlagSet) (Factory, error) {
	if clientCtx.Viper == nil {
		clientCtx = clientCtx.WithViper("")
	}

	if err := clientCtx.Viper.BindPFlags(flagSet); err != nil {
		return Factory{}, fmt.Errorf("failed to bind flags to viper: %w", err)
	}

	var accNum, accSeq uint64
	if clientCtx.Offline {
		if flagSet.Changed(flags.FlagAccountNumber) && flagSet.Changed(flags.FlagSequence) {
			accNum = clientCtx.Viper.GetUint64(flags.FlagAccountNumber)
			accSeq = clientCtx.Viper.GetUint64(flags.FlagSequence)
		} else {
			return Factory{}, fmt.Errorf("account-number and sequence must be set in offline mode")
		}
	}

	if clientCtx.Offline && clientCtx.GenerateOnly {
		if clientCtx.ChainID != "" {
			return Factory{}, errors.New("chain ID cannot be used when offline and generate-only flags are set")
		}
	} else if clientCtx.ChainID == "" {
		return Factory{}, errors.New("chain ID required but not specified")
	}

	signMode := flags.ParseSignMode(clientCtx.SignModeStr)
	memo := clientCtx.Viper.GetString(flags.FlagNote)
	timeoutHeight := clientCtx.Viper.GetUint64(flags.FlagTimeoutHeight)
	unordered := clientCtx.Viper.GetBool(flags.FlagUnordered)

	gasAdj := clientCtx.Viper.GetFloat64(flags.FlagGasAdjustment)
	gasStr := clientCtx.Viper.GetString(flags.FlagGas)
	gasSetting, _ := flags.ParseGasSetting(gasStr)
	gasPricesStr := clientCtx.Viper.GetString(flags.FlagGasPrices)

	feesStr := clientCtx.Viper.GetString(flags.FlagFees)

	f := Factory{
		accountRetriever: clientCtx.AccountRetriever,
		keybase:          clientCtx.Keyring,
		TxParameters: TxParameters{
			timeoutHeight: timeoutHeight,
			memo:          memo,
			chainID:       clientCtx.ChainID,
			signMode:      signMode,
			AccountConfig: AccountConfig{
				accountNumber: accNum,
				sequence:      accSeq,
				fromName:      clientCtx.FromName,
				fromAddress:   clientCtx.FromAddress,
			},
			GasConfig: GasConfig{
				gas:           gasSetting.Gas,
				gasAdjustment: gasAdj,
			},
			FeeConfig: FeeConfig{
				feeGranter: clientCtx.FeeGranter,
				feePayer:   clientCtx.FeePayer,
			},
			ExecutionOptions: ExecutionOptions{
				unordered:          unordered,
				offline:            clientCtx.Offline,
				generateOnly:       clientCtx.GenerateOnly,
				simulateAndExecute: gasSetting.Simulate,
				preprocessTxHook:   clientCtx.PreprocessTxHook,
			},
		},
	}

	f = f.WithFees(feesStr).WithGasPrices(gasPricesStr)
	return f, nil
}

// Prepare ensures the account defined by ctx.GetFromAddress() exists and
// if the account number and/or the account sequence number are zero (not set),
// they will be queried for and set on the provided Factory.
// A new Factory with the updated fields will be returned.
// Note: When in offline mode, the Prepare does nothing and returns the original factory.
func (f Factory) Prepare(clientCtx txContext) (Factory, error) {
	if f.ExecutionOptions.offline {
		return f, nil
	}

	if f.fromAddress.Empty() {
		return f, errors.New("missing 'from address' field")
	}

	if err := f.accountRetriever.EnsureExists(clientCtx, f.fromAddress); err != nil {
		return f, err
	}

	if f.accountNumber == 0 || f.sequence == 0 {
		fc := f
		num, seq, err := fc.accountRetriever.GetAccountNumberSequence(clientCtx, f.fromAddress)
		if err != nil {
			return f, err
		}

		if f.accountNumber == 0 {
			fc = fc.WithAccountNumber(num)
		}

		if f.sequence == 0 {
			fc = fc.WithSequence(seq)
		}

		return fc, nil
	}

	return f, nil
}

// BuildUnsignedTx builds a transaction to be signed given a set of messages.
// Once created, the fee, memo, and messages are set.
func (f Factory) BuildUnsignedTx(msgs ...sdk.MsgV2) (TxBuilder, error) {
	fees := f.fees

	if !f.gasPrices.IsZero() {
		if !fees.IsZero() {
			return nil, errors.New("cannot provide both fees and gas prices")
		}

		// f.gas is a uint64 and we should convert to LegacyDec
		// without the risk of under/overflow via uint64->int64.
		glDec := math.LegacyNewDecFromBigInt(new(big.Int).SetUint64(f.gas))

		// Derive the fees based on the provided gas prices, where
		// fee = ceil(gasPrice * gasLimit).
		fees = make(sdk.Coins, len(f.gasPrices))

		for i, gp := range f.gasPrices {
			fee := gp.Amount.Mul(glDec)
			fees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	// Prevent simple inclusion of a valid mnemonic in the memo field
	if f.memo != "" && bip39.IsMnemonicValid(strings.ToLower(f.memo)) {
		return nil, errors.New("cannot provide a valid mnemonic seed in the memo field")
	}

	txBuilder := f.txBuilderProvider.NewTxBuilder()

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	txBuilder.SetMemo(f.memo)
	txBuilder.SetFeeAmount(fees)
	txBuilder.SetGasLimit(f.gas)
	txBuilder.SetFeeGranter(f.feeGranter)
	txBuilder.SetFeePayer(f.feePayer)
	txBuilder.SetTimeoutHeight(f.timeoutHeight)

	if etx, ok := txBuilder.(ExtendedTxBuilder); ok {
		etx.SetExtensionOptions(f.extOptions...)
	}

	return txBuilder, nil
}

// PrintUnsignedTx will generate an unsigned transaction and print it to the writer
// specified by ctx.Output. If simulation was requested, the gas will be
// simulated and also printed to the same writer before the transaction is
// printed.
func (f Factory) PrintUnsignedTx(clientCtx txContext, msgs ...sdk.MsgV2) error {
	if f.SimulateAndExecute() {
		if clientCtx.Offline {
			return errors.New("cannot estimate gas in offline mode")
		}

		// Prepare TxFactory with acc & seq numbers as CalculateGas requires
		// account and sequence numbers to be set
		preparedTxf, err := f.Prepare(clientCtx)
		if err != nil {
			return err
		}

		_, adjusted, err := CalculateGas(clientCtx, preparedTxf, msgs...)
		if err != nil {
			return err
		}

		f = f.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: f.Gas()})
	}

	unsignedTx, err := f.BuildUnsignedTx(msgs...)
	if err != nil {
		return err
	}

	encoder := f.codec.TxJSONEncoder()
	if encoder == nil {
		return errors.New("cannot print unsigned tx: tx json encoder is nil")
	}

	json, err := encoder(unsignedTx.GetTx())
	if err != nil {
		return err
	}

	return clientCtx.PrintString(fmt.Sprintf("%s\n", json))
}

// BuildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be
// built.
func (f Factory) BuildSimTx(msgs ...sdk.MsgV2) ([]byte, error) {
	txb, err := f.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	pk, err := f.getSimPK()
	if err != nil {
		return nil, err
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		PubKey:   pk,
		Data:     f.getSimSignatureData(pk),
		Sequence: f.Sequence(),
	}
	if err := txb.SetSignatures(sig); err != nil {
		return nil, err
	}

	encoder := f.txConfig.TxEncoder()
	if encoder == nil {
		return nil, fmt.Errorf("cannot simulate tx: tx encoder is nil")
	}

	return encoder(txb.GetTx())
}

// WithTxConfig returns a copy of the Factory with an updated TxConfig.
func (f Factory) WithTxConfig(g TxConfig) Factory {
	f.txConfig = g
	return f
}

// WithAccountRetriever returns a copy of the Factory with an updated AccountRetriever.
func (f Factory) WithAccountRetriever(ar AccountRetriever) Factory {
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

// WithFromName returns a copy of the Factory with updated fromName
// fromName will be use for building a simulation tx.
func (f Factory) WithFromName(fromName string) Factory {
	f.fromName = fromName
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

// WithFeeGranter returns a copy of the Factory with an updated fee granter.
func (f Factory) WithFeeGranter(fg sdk.AccAddress) Factory {
	f.feeGranter = fg
	return f
}

// WithFeePayer returns a copy of the Factory with an updated fee granter.
func (f Factory) WithFeePayer(fp sdk.AccAddress) Factory {
	f.feePayer = fp
	return f
}

// WithPreprocessTxHook returns a copy of the Factory with an updated preprocess tx function,
// allows for preprocessing of transaction data using the TxBuilder.
func (f Factory) WithPreprocessTxHook(preprocessFn PreprocessTxFn) Factory {
	f.preprocessTxHook = preprocessFn
	return f
}

func (f Factory) WithExtensionOptions(extOpts ...*codectypes.Any) Factory {
	f.extOptions = extOpts
	return f
}

// PreprocessTx calls the preprocessing hook with the factory parameters and
// returns the result.
func (f Factory) PreprocessTx(keyname string, builder TxBuilder) error {
	if f.preprocessTxHook == nil {
		// Allow pass-through
		return nil
	}

	key, err := f.Keybase().Key(keyname)
	if err != nil {
		return fmt.Errorf("error retrieving key from keyring: %w", err)
	}

	return f.preprocessTxHook(f.chainID, f.Keybase().GetRecordType(key), builder)
}

func (f Factory) AccountNumber() uint64              { return f.accountNumber }
func (f Factory) Sequence() uint64                   { return f.sequence }
func (f Factory) Gas() uint64                        { return f.gas }
func (f Factory) GasAdjustment() float64             { return f.gasAdjustment }
func (f Factory) Keybase() keyring.Keyring           { return f.keybase }
func (f Factory) ChainID() string                    { return f.chainID }
func (f Factory) Memo() string                       { return f.memo }
func (f Factory) Fees() sdk.Coins                    { return f.fees }
func (f Factory) GasPrices() sdk.DecCoins            { return f.gasPrices }
func (f Factory) AccountRetriever() AccountRetriever { return f.accountRetriever }
func (f Factory) TimeoutHeight() uint64              { return f.timeoutHeight }
func (f Factory) FromName() string                   { return f.fromName }
func (f Factory) SimulateAndExecute() bool           { return f.simulateAndExecute }
func (f Factory) SignMode() signing.SignMode         { return f.signMode }

// getSimPK gets the public key to use for building a simulation tx.
// Note, we should only check for keys in the keybase if we are in simulate and execute mode,
// e.g. when using --gas=auto.
// When using --dry-run, we are is simulation mode only and should not check the keybase.
// Ref: https://github.com/cosmos/cosmos-sdk/issues/11283
func (f Factory) getSimPK() (cryptotypes.PubKey, error) {
	var (
		err error
		pk  cryptotypes.PubKey = &secp256k1.PubKey{} // use default public key type
	)

	if f.simulateAndExecute && f.keybase != nil {
		pk, err = f.keybase.GetPubKey(f.fromName)
		if err != nil {
			return nil, err
		}
	}

	return pk, nil
}

// getSimSignatureData based on the pubKey type gets the correct SignatureData type
// to use for building a simulation tx.
func (f Factory) getSimSignatureData(pk cryptotypes.PubKey) signing.SignatureData {
	multisigPubKey, ok := pk.(*multisig.LegacyAminoPubKey) // TODO: abstract out multisig pubkey
	if !ok {
		return &signing.SingleSignatureData{SignMode: f.signMode}
	}

	multiSignatureData := make([]signing.SignatureData, 0, multisigPubKey.Threshold)
	for i := uint32(0); i < multisigPubKey.Threshold; i++ {
		multiSignatureData = append(multiSignatureData, &signing.SingleSignatureData{
			SignMode: f.SignMode(),
		})
	}

	return &signing.MultiSignatureData{
		Signatures: multiSignatureData,
	}
}

// Sign signs a given tx with a named key. The bytes signed over are canconical.
// The resulting signature will be added to the transaction builder overwriting the previous
// ones if overwrite=true (otherwise, the signature will be appended).
// Signing a transaction with mutltiple signers in the DIRECT mode is not supported and will
// return an error.
// An error is returned upon failure.
func Sign(ctx context.Context, txf Factory, name string, txBuilder TxBuilder, overwriteSig bool) error {
	if txf.keybase == nil {
		return errors.New("keybase must be set prior to signing a transaction")
	}

	var err error
	signMode := txf.signMode
	if signMode == signing.SignMode_SIGN_MODE_UNSPECIFIED {
		// use the SignModeHandler's default mode if unspecified
		signMode, err = authsigning.APISignModeToInternal(txf.txConfig.SignModeHandler().DefaultMode())
		if err != nil {
			return err
		}
	}

	pubKey, err := txf.keybase.GetPubKey(name)
	if err != nil {
		return err
	}

	signerData := authsigning.SignerData{
		ChainID:       txf.chainID,
		AccountNumber: txf.accountNumber,
		Sequence:      txf.sequence,
		PubKey:        pubKey,
		Address:       sdk.AccAddress(pubKey.Address()).String(),
	}

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generated the
	// sign bytes. This is the reason for setting SetSignatures here, with a
	// nil signature.
	//
	// Note: this line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	sigData := signing.SingleSignatureData{
		SignMode:  txf.signMode,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}

	var prevSignatures []signing.SignatureV2
	if !overwriteSig {
		prevSignatures, err = txBuilder.GetTx().GetSignaturesV2()
		if err != nil {
			return err
		}
	}
	// Overwrite or append signer infos.
	var sigs []signing.SignatureV2
	if overwriteSig {
		sigs = []signing.SignatureV2{sig}
	} else {
		sigs = append(sigs, prevSignatures...)
		sigs = append(sigs, sig)
	}
	if err := txBuilder.SetSignatures(sigs...); err != nil {
		return err
	}

	if err := checkMultipleSigners(txBuilder.GetTx()); err != nil {
		return err
	}

	bytesToSign, err := authsigning.GetSignBytesAdapter(ctx, txf.txConfig.SignModeHandler(), signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return err
	}

	// Sign those bytes
	sigBytes, err := txf.keybase.Sign(name, bytesToSign, signMode)
	if err != nil {
		return err
	}

	// Construct the SignatureV2 struct
	sigData = signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: sigBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}

	if overwriteSig {
		err = txBuilder.SetSignatures(sig)
	} else {
		prevSignatures = append(prevSignatures, sig)
		err = txBuilder.SetSignatures(prevSignatures...)
	}

	if err != nil {
		return fmt.Errorf("unable to set signatures on payload: %w", err)
	}

	// Run optional preprocessing if specified. By default, this is unset
	// and will return nil.
	return txf.PreprocessTx(name, txBuilder)
}
