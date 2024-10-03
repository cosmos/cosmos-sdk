package tx

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/cosmos/go-bip39"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apicrypto "cosmossdk.io/api/cosmos/crypto/multisig/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/client/v2/internal/account"
	"cosmossdk.io/client/v2/internal/coins"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	"cosmossdk.io/x/tx/signing"

	flags2 "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// Factory defines a client transaction factory that facilitates generating and
// signing an application-specific transaction.
type Factory struct {
	keybase          keyring.Keyring
	cdc              codec.BinaryCodec
	accountRetriever account.AccountRetriever
	ac               address.Codec
	conn             gogogrpc.ClientConn
	txConfig         TxConfig
	txParams         TxParameters

	tx txState
}

func NewFactoryFromFlagSet(flags *pflag.FlagSet, keybase keyring.Keyring, cdc codec.BinaryCodec, accRetriever account.AccountRetriever,
	txConfig TxConfig, ac address.Codec, conn gogogrpc.ClientConn,
) (Factory, error) {
	offline, _ := flags.GetBool(flags2.FlagOffline)
	if err := validateFlagSet(flags, offline); err != nil {
		return Factory{}, err
	}

	params, err := txParamsFromFlagSet(flags, keybase, ac)
	if err != nil {
		return Factory{}, err
	}

	params, err = prepareTxParams(params, accRetriever, offline)
	if err != nil {
		return Factory{}, err
	}

	return NewFactory(keybase, cdc, accRetriever, txConfig, ac, conn, params)
}

// NewFactory returns a new instance of Factory.
func NewFactory(keybase keyring.Keyring, cdc codec.BinaryCodec, accRetriever account.AccountRetriever,
	txConfig TxConfig, ac address.Codec, conn gogogrpc.ClientConn, parameters TxParameters,
) (Factory, error) {
	return Factory{
		keybase:          keybase,
		cdc:              cdc,
		accountRetriever: accRetriever,
		ac:               ac,
		conn:             conn,
		txConfig:         txConfig,
		txParams:         parameters,

		tx: txState{},
	}, nil
}

// validateFlagSet checks the provided flags for consistency and requirements based on the operation mode.
func validateFlagSet(flags *pflag.FlagSet, offline bool) error {
	if offline {
		if !flags.Changed(flags2.FlagAccountNumber) || !flags.Changed(flags2.FlagSequence) {
			return errors.New("account-number and sequence must be set in offline mode")
		}

		gas, _ := flags.GetString(flags2.FlagGas)
		gasSetting, _ := flags2.ParseGasSetting(gas)
		if gasSetting.Simulate {
			return errors.New("simulate and offline flags cannot be set at the same time")
		}
	}

	generateOnly, _ := flags.GetBool(flags2.FlagGenerateOnly)
	chainID, _ := flags.GetString(flags2.FlagChainID)
	if offline && generateOnly && chainID != "" {
		return errors.New("chain ID cannot be used when offline and generate-only flags are set")
	}
	if chainID == "" {
		return errors.New("chain ID required but not specified")
	}

	dryRun, _ := flags.GetBool(flags2.FlagDryRun)
	if offline && dryRun {
		return errors.New("dry-run: cannot use offline mode")
	}

	return nil
}

// prepareTxParams ensures the account defined by ctx.GetFromAddress() exists and
// if the account number and/or the account sequence number are zero (not set),
// they will be queried for and set on the provided Factory.
func prepareTxParams(parameters TxParameters, accRetriever account.AccountRetriever, offline bool) (TxParameters, error) {
	if offline {
		return parameters, nil
	}

	if len(parameters.address) == 0 {
		return parameters, errors.New("missing 'from address' field")
	}

	if parameters.accountNumber == 0 || parameters.sequence == 0 {
		num, seq, err := accRetriever.GetAccountNumberSequence(context.Background(), parameters.address)
		if err != nil {
			return parameters, err
		}

		if parameters.accountNumber == 0 {
			parameters.accountNumber = num
		}

		if parameters.sequence == 0 {
			parameters.sequence = seq
		}
	}

	return parameters, nil
}

// BuildUnsignedTx builds a transaction to be signed given a set of messages.
// Once created, the fee, memo, and messages are set.
func (f *Factory) BuildUnsignedTx(msgs ...transaction.Msg) error {
	fees := f.txParams.fees

	isGasPriceZero, err := coins.IsZero(f.txParams.gasPrices)
	if err != nil {
		return err
	}
	if !isGasPriceZero {
		areFeesZero, err := coins.IsZero(fees)
		if err != nil {
			return err
		}
		if !areFeesZero {
			return errors.New("cannot provide both fees and gas prices")
		}

		// f.gas is an uint64 and we should convert to LegacyDec
		// without the risk of under/overflow via uint64->int64.
		glDec := math.LegacyNewDecFromBigInt(new(big.Int).SetUint64(f.txParams.gas))

		// Derive the fees based on the provided gas prices, where
		// fee = ceil(gasPrice * gasLimit).
		fees = make([]*base.Coin, len(f.txParams.gasPrices))

		for i, gp := range f.txParams.gasPrices {
			fee, err := math.LegacyNewDecFromStr(gp.Amount)
			if err != nil {
				return err
			}
			fee = fee.Mul(glDec)
			fees[i] = &base.Coin{Denom: gp.Denom, Amount: fee.Ceil().RoundInt().String()}
		}
	}

	if err := validateMemo(f.txParams.memo); err != nil {
		return err
	}

	f.tx.msgs = msgs
	f.tx.memo = f.txParams.memo
	f.tx.fees = fees
	f.tx.gasLimit = f.txParams.gas
	f.tx.unordered = f.txParams.unordered
	f.tx.timeoutTimestamp = f.txParams.timeoutTimestamp

	err = f.setFeeGranter(f.txParams.feeGranter)
	if err != nil {
		return err
	}
	err = f.setFeePayer(f.txParams.feePayer)
	if err != nil {
		return err
	}

	return nil
}

func (f *Factory) BuildsSignedTx(ctx context.Context, msgs ...transaction.Msg) (Tx, error) {
	err := f.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	return f.sign(ctx, true)
}

// calculateGas calculates the gas required for the given messages.
func (f *Factory) calculateGas(msgs ...transaction.Msg) error {
	_, adjusted, err := f.Simulate(msgs...)
	if err != nil {
		return err
	}

	f.WithGas(adjusted)

	return nil
}

// Simulate simulates the execution of a transaction and returns the
// simulation response obtained by the query and the adjusted gas amount.
func (f *Factory) Simulate(msgs ...transaction.Msg) (*apitx.SimulateResponse, uint64, error) {
	txBytes, err := f.BuildSimTx(msgs...)
	if err != nil {
		return nil, 0, err
	}

	txSvcClient := apitx.NewServiceClient(f.conn)
	simRes, err := txSvcClient.Simulate(context.Background(), &apitx.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, 0, err
	}

	return simRes, uint64(f.gasAdjustment() * float64(simRes.GasInfo.GasUsed)), nil
}

// UnsignedTxString will generate an unsigned transaction and print it to the writer
// specified by ctx.Output. If simulation was requested, the gas will be
// simulated and also printed to the same writer before the transaction is
// printed.
func (f *Factory) UnsignedTxString(msgs ...transaction.Msg) (string, error) {
	if f.simulateAndExecute() {
		err := f.calculateGas(msgs...)
		if err != nil {
			return "", err
		}
	}

	err := f.BuildUnsignedTx(msgs...)
	if err != nil {
		return "", err
	}

	encoder := f.txConfig.TxJSONEncoder()
	if encoder == nil {
		return "", errors.New("cannot print unsigned tx: tx json encoder is nil")
	}

	tx, err := f.getTx()
	if err != nil {
		return "", err
	}

	json, err := encoder(tx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n", json), nil
}

// BuildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be
// built.
func (f *Factory) BuildSimTx(msgs ...transaction.Msg) ([]byte, error) {
	err := f.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	pk, err := f.getSimPK()
	if err != nil {
		return nil, err
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := Signature{
		PubKey:   pk,
		Data:     f.getSimSignatureData(pk),
		Sequence: f.sequence(),
	}
	if err := f.setSignatures(sig); err != nil {
		return nil, err
	}

	encoder := f.txConfig.TxEncoder()
	if encoder == nil {
		return nil, fmt.Errorf("cannot simulate tx: tx encoder is nil")
	}

	tx, err := f.getTx()
	if err != nil {
		return nil, err
	}
	return encoder(tx)
}

// sign signs a given tx with a named key. The bytes signed over are canonical.
// The resulting signature will be added to the transaction builder overwriting the previous
// ones if overwrite=true (otherwise, the signature will be appended).
// Signing a transaction with multiple signers in the DIRECT mode is not supported and will
// return an error.
func (f *Factory) sign(ctx context.Context, overwriteSig bool) (Tx, error) {
	if f.keybase == nil {
		return nil, errors.New("keybase must be set prior to signing a transaction")
	}

	var err error
	if f.txParams.signMode == apitxsigning.SignMode_SIGN_MODE_UNSPECIFIED {
		f.txParams.signMode = f.txConfig.SignModeHandler().DefaultMode()
	}

	pubKey, err := f.keybase.GetPubKey(f.txParams.fromName)
	if err != nil {
		return nil, err
	}

	addr, err := f.ac.BytesToString(pubKey.Address())
	if err != nil {
		return nil, err
	}

	signerData := signing.SignerData{
		ChainID:       f.txParams.chainID,
		AccountNumber: f.txParams.accountNumber,
		Sequence:      f.txParams.sequence,
		PubKey: &anypb.Any{
			TypeUrl: codectypes.MsgTypeURL(pubKey),
			Value:   pubKey.Bytes(),
		},
		Address: addr,
	}

	// For SIGN_MODE_DIRECT, we need to set the SignerInfos before generating
	// the sign bytes. This is done by calling setSignatures with a nil
	// signature, which in turn calls setSignerInfos internally.
	//
	// For SIGN_MODE_LEGACY_AMINO, this step is not strictly necessary,
	// but we include it for consistency across all sign modes.
	// It does not affect the generated sign bytes for LEGACY_AMINO.
	//
	// By setting the signatures here, we ensure that the correct SignerInfos
	// are in place for all subsequent operations, regardless of the sign mode.
	sigData := SingleSignatureData{
		SignMode:  f.txParams.signMode,
		Signature: nil,
	}
	sig := Signature{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: f.txParams.sequence,
	}

	var prevSignatures []Signature
	if !overwriteSig {
		tx, err := f.getTx()
		if err != nil {
			return nil, err
		}

		prevSignatures, err = tx.GetSignatures()
		if err != nil {
			return nil, err
		}
	}
	// Overwrite or append signer infos.
	var sigs []Signature
	if overwriteSig {
		sigs = []Signature{sig}
	} else {
		sigs = append(sigs, prevSignatures...)
		sigs = append(sigs, sig)
	}
	if err := f.setSignatures(sigs...); err != nil {
		return nil, err
	}

	tx, err := f.getTx()
	if err != nil {
		return nil, err
	}

	if err := checkMultipleSigners(tx); err != nil {
		return nil, err
	}

	bytesToSign, err := f.getSignBytesAdapter(ctx, signerData)
	if err != nil {
		return nil, err
	}

	// Sign those bytes
	sigBytes, err := f.keybase.Sign(f.txParams.fromName, bytesToSign, f.txParams.signMode)
	if err != nil {
		return nil, err
	}

	// Construct the SignatureV2 struct
	sigData = SingleSignatureData{
		SignMode:  f.signMode(),
		Signature: sigBytes,
	}
	sig = Signature{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: f.txParams.sequence,
	}

	if overwriteSig {
		err = f.setSignatures(sig)
	} else {
		prevSignatures = append(prevSignatures, sig)
		err = f.setSignatures(prevSignatures...)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to set signatures on payload: %w", err)
	}

	return f.getTx()
}

// getSignBytesAdapter returns the sign bytes for a given transaction and sign mode.
func (f *Factory) getSignBytesAdapter(ctx context.Context, signerData signing.SignerData) ([]byte, error) {
	txData, err := f.getSigningTxData()
	if err != nil {
		return nil, err
	}

	// Generate the bytes to be signed.
	return f.txConfig.SignModeHandler().GetSignBytes(ctx, f.signMode(), signerData, *txData)
}

// WithGas returns a copy of the Factory with an updated gas value.
func (f *Factory) WithGas(gas uint64) {
	f.txParams.gas = gas
}

// WithSequence returns a copy of the Factory with an updated sequence number.
func (f *Factory) WithSequence(sequence uint64) {
	f.txParams.sequence = sequence
}

// WithAccountNumber returns a copy of the Factory with an updated account number.
func (f *Factory) WithAccountNumber(accnum uint64) {
	f.txParams.accountNumber = accnum
}

// sequence returns the sequence number.
func (f *Factory) sequence() uint64 { return f.txParams.sequence }

// gasAdjustment returns the gas adjustment value.
func (f *Factory) gasAdjustment() float64 { return f.txParams.gasAdjustment }

// simulateAndExecute returns whether to simulate and execute.
func (f *Factory) simulateAndExecute() bool { return f.txParams.simulateAndExecute }

// signMode returns the sign mode.
func (f *Factory) signMode() apitxsigning.SignMode { return f.txParams.signMode }

// getSimPK gets the public key to use for building a simulation tx.
// Note, we should only check for keys in the keybase if we are in simulate and execute mode,
// e.g. when using --gas=auto.
// When using --dry-run, we are is simulation mode only and should not check the keybase.
// Ref: https://github.com/cosmos/cosmos-sdk/issues/11283
func (f *Factory) getSimPK() (cryptotypes.PubKey, error) {
	var (
		err error
		pk  cryptotypes.PubKey = &secp256k1.PubKey{}
	)

	if f.txParams.simulateAndExecute && f.keybase != nil {
		pk, err = f.keybase.GetPubKey(f.txParams.fromName)
		if err != nil {
			return nil, err
		}
	} else {
		// When in dry-run mode, attempt to retrieve the account using the provided address.
		// If the account retrieval fails, the default public key is used.
		acc, err := f.accountRetriever.GetAccount(context.Background(), f.txParams.address)
		if err != nil {
			// If there is an error retrieving the account, return the default public key.
			return pk, nil
		}
		// If the account is successfully retrieved, use its public key.
		pk = acc.GetPubKey()
	}

	return pk, nil
}

// getSimSignatureData based on the pubKey type gets the correct SignatureData type
// to use for building a simulation tx.
func (f *Factory) getSimSignatureData(pk cryptotypes.PubKey) SignatureData {
	multisigPubKey, ok := pk.(*multisig.LegacyAminoPubKey)
	if !ok {
		return &SingleSignatureData{SignMode: f.txParams.signMode}
	}

	multiSignatureData := make([]SignatureData, 0, multisigPubKey.Threshold)
	for i := uint32(0); i < multisigPubKey.Threshold; i++ {
		multiSignatureData = append(multiSignatureData, &SingleSignatureData{
			SignMode: f.signMode(),
		})
	}

	return &MultiSignatureData{
		BitArray:   &apicrypto.CompactBitArray{},
		Signatures: multiSignatureData,
	}
}

func (f *Factory) getTx() (*wrappedTx, error) {
	msgs, err := msgsV1toAnyV2(f.tx.msgs)
	if err != nil {
		return nil, err
	}

	body := &apitx.TxBody{
		Messages:                    msgs,
		Memo:                        f.tx.memo,
		TimeoutHeight:               f.tx.timeoutHeight,
		TimeoutTimestamp:            timestamppb.New(f.tx.timeoutTimestamp),
		Unordered:                   f.tx.unordered,
		ExtensionOptions:            f.tx.extensionOptions,
		NonCriticalExtensionOptions: f.tx.nonCriticalExtensionOptions,
	}

	fee, err := f.getFee()
	if err != nil {
		return nil, err
	}

	authInfo := &apitx.AuthInfo{
		SignerInfos: f.tx.signerInfos,
		Fee:         fee,
	}

	bodyBytes, err := marshalOption.Marshal(body)
	if err != nil {
		return nil, err
	}

	authInfoBytes, err := marshalOption.Marshal(authInfo)
	if err != nil {
		return nil, err
	}

	txRawBytes, err := marshalOption.Marshal(&apitx.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		Signatures:    f.tx.signatures,
	})
	if err != nil {
		return nil, err
	}

	decodedTx, err := f.txConfig.Decoder().Decode(txRawBytes)
	if err != nil {
		return nil, err
	}

	return newWrapperTx(f.cdc, decodedTx), nil
}

// getSigningTxData returns a TxData with the current txState info.
func (f *Factory) getSigningTxData() (*signing.TxData, error) {
	tx, err := f.getTx()
	if err != nil {
		return nil, err
	}

	return &signing.TxData{
		Body:                       tx.Tx.Body,
		AuthInfo:                   tx.Tx.AuthInfo,
		BodyBytes:                  tx.TxRaw.BodyBytes,
		AuthInfoBytes:              tx.TxRaw.AuthInfoBytes,
		BodyHasUnknownNonCriticals: tx.TxBodyHasUnknownNonCriticals,
	}, nil
}

// setSignatures sets the signatures for the transaction builder.
// It takes a variable number of Signature arguments and processes each one to extract the mode information and raw signature.
// It also converts the public key to the appropriate format and sets the signer information.
func (f *Factory) setSignatures(signatures ...Signature) error {
	n := len(signatures)
	signerInfos := make([]*apitx.SignerInfo, n)
	rawSignatures := make([][]byte, n)

	for i, sig := range signatures {
		var (
			modeInfo *apitx.ModeInfo
			pubKey   *codectypes.Any
			err      error
			anyPk    *anypb.Any
		)

		modeInfo, rawSignatures[i] = signatureDataToModeInfoAndSig(sig.Data)
		if sig.PubKey != nil {
			pubKey, err = codectypes.NewAnyWithValue(sig.PubKey)
			if err != nil {
				return err
			}
			anyPk = &anypb.Any{
				TypeUrl: pubKey.TypeUrl,
				Value:   pubKey.Value,
			}
		}

		signerInfos[i] = &apitx.SignerInfo{
			PublicKey: anyPk,
			ModeInfo:  modeInfo,
			Sequence:  sig.Sequence,
		}
	}

	f.tx.signerInfos = signerInfos
	f.tx.signatures = rawSignatures

	return nil
}

// getFee computes the transaction fee information.
// It returns a pointer to an apitx.Fee struct containing the fee amount, gas limit, payer, and granter information.
// If the granter or payer addresses are set, it converts them from bytes to string using the addressCodec.
func (f *Factory) getFee() (fee *apitx.Fee, err error) {
	granterStr := ""
	if f.tx.granter != nil {
		granterStr, err = f.ac.BytesToString(f.tx.granter)
		if err != nil {
			return nil, err
		}
	}

	payerStr := ""
	if f.tx.payer != nil {
		payerStr, err = f.ac.BytesToString(f.tx.payer)
		if err != nil {
			return nil, err
		}
	}

	fee = &apitx.Fee{
		Amount:   f.tx.fees,
		GasLimit: f.tx.gasLimit,
		Payer:    payerStr,
		Granter:  granterStr,
	}

	return fee, nil
}

// setFeePayer sets the fee payer for the transaction.
func (f *Factory) setFeePayer(feePayer string) error {
	if feePayer == "" {
		return nil
	}

	addr, err := f.ac.StringToBytes(feePayer)
	if err != nil {
		return err
	}
	f.tx.payer = addr
	return nil
}

// setFeeGranter sets the fee granter's address in the transaction builder.
// If the feeGranter string is empty, the function returns nil without setting an address.
// It converts the feeGranter string to bytes using the address codec and sets it as the granter address.
// Returns an error if the conversion fails.
func (f *Factory) setFeeGranter(feeGranter string) error {
	if feeGranter == "" {
		return nil
	}

	addr, err := f.ac.StringToBytes(feeGranter)
	if err != nil {
		return err
	}
	f.tx.granter = addr

	return nil
}

// msgsV1toAnyV2 converts a slice of transaction.Msg (v1) to a slice of anypb.Any (v2).
// It first converts each transaction.Msg into a codectypes.Any and then converts
// these into anypb.Any.
func msgsV1toAnyV2(msgs []transaction.Msg) ([]*anypb.Any, error) {
	anys := make([]*codectypes.Any, len(msgs))
	for i, msg := range msgs {
		anyMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
		anys[i] = anyMsg
	}

	return intoAnyV2(anys), nil
}

// intoAnyV2 converts a slice of codectypes.Any (v1) to a slice of anypb.Any (v2).
func intoAnyV2(v1s []*codectypes.Any) []*anypb.Any {
	v2s := make([]*anypb.Any, len(v1s))
	for i, v1 := range v1s {
		v2s[i] = &anypb.Any{
			TypeUrl: v1.TypeUrl,
			Value:   v1.Value,
		}
	}
	return v2s
}

// checkMultipleSigners checks that there can be maximum one DIRECT signer in
// a tx.
func checkMultipleSigners(tx Tx) error {
	directSigners := 0
	sigsV2, err := tx.GetSignatures()
	if err != nil {
		return err
	}
	for _, sig := range sigsV2 {
		directSigners += countDirectSigners(sig.Data)
		if directSigners > 1 {
			return errors.New("txs signed with CLI can have maximum 1 DIRECT signer")
		}
	}

	return nil
}

// validateMemo validates the memo field.
func validateMemo(memo string) error {
	// Prevent simple inclusion of a valid mnemonic in the memo field
	if memo != "" && bip39.IsMnemonicValid(strings.ToLower(memo)) {
		return errors.New("cannot provide a valid mnemonic seed in the memo field")
	}

	return nil
}
