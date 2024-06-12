package tx

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/spf13/pflag"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	keyring2 "cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/client"
	flags2 "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

func txParamsFromFlagSet(flags *pflag.FlagSet, keybase keyring2.Keyring) (params TxParameters, err error) {
	timeout, _ := flags.GetUint64(flags2.FlagTimeoutHeight)
	chainID, _ := flags.GetString(flags2.FlagChainID)
	memo, _ := flags.GetString(flags2.FlagNote)
	signMode, _ := flags.GetString(flags2.FlagSignMode)
	accNumber, _ := flags.GetUint64(flags2.FlagAccountNumber)
	sequence, _ := flags.GetUint64(flags2.FlagSequence)
	fromName, _ := flags.GetString(flags2.FlagFrom)
	gas, _ := flags.GetUint64(flags2.FlagGas)
	gasAdjustment, _ := flags.GetFloat64(flags2.FlagGasAdjustment)
	gasPrices, _ := flags.GetString(flags2.FlagGasPrices)

	fees, _ := flags.GetString(flags2.FlagFees)
	feePayer, _ := flags.GetString(flags2.FlagFeePayer)
	feeGrater, _ := flags.GetString(flags2.FlagFeeGranter)

	unordered, _ := flags.GetBool(flags2.FlagUnordered)
	offline, _ := flags.GetBool(flags2.FlagOffline)
	generateOnly, _ := flags.GetBool(flags2.FlagGenerateOnly)

	acc, err := keybase.GetPubKey(fromName)
	if err != nil {
		return params, err
	}

	gasConfig, err := NewGasConfig(gas, gasAdjustment, gasPrices)
	if err != nil {
		return params, err
	}
	feeConfig, err := NewFeeConfig(fees, feePayer, feeGrater)
	if err != nil {
		return params, err
	}

	txParams := TxParameters{
		timeoutHeight: timeout,
		chainID:       chainID,
		memo:          memo,
		signMode:      getSignMode(signMode),
		AccountConfig: AccountConfig{
			accountNumber: accNumber,
			sequence:      sequence,
			fromName:      fromName,
			fromAddress:   acc.Address().Bytes(),
		},
		GasConfig: gasConfig,
		FeeConfig: feeConfig,
		ExecutionOptions: ExecutionOptions{
			unordered:          unordered,
			offline:            offline,
			offChain:           false,
			generateOnly:       generateOnly,
			simulateAndExecute: false, // TODO: in context
			preprocessTxHook:   nil,   // TODO: in context
		},
		ExtensionOptions: ExtensionOptions{}, // TODO
	}

	return txParams, nil
}

func validate(flags *pflag.FlagSet) error {
	offline, _ := flags.GetBool(flags2.FlagOffline)
	if offline {
		if !flags.Changed(flags2.FlagAccountNumber) || !flags.Changed(flags2.FlagSequence) {
			return errors.New("account-number and sequence must be set in offline mode")
		}
	}

	generateOnly, _ := flags.GetBool(flags2.FlagGenerateOnly)
	chainID, _ := flags.GetString(flags2.FlagChainID)
	if offline && generateOnly {
		if chainID != "" {
			return errors.New("chain ID cannot be used when offline and generate-only flags are set")
		}
	}
	if chainID == "" {
		return errors.New("chain ID required but not specified")
	}

	return nil
}

// GenerateOrBroadcastTxCLI will either generate and print an unsigned transaction
// or sign it and broadcast it returning an error upon failure.
// TODO: remove the client.Context
func GenerateOrBroadcastTxCLI(ctx client.Context, flagSet *pflag.FlagSet, msgs ...transaction.Msg) error {
	if err := validate(flagSet); err != nil {
		return err
	}

	txf, err := newFactory(ctx, flagSet, msgs...)
	if err != nil {
		return err
	}

	isAux, _ := flagSet.GetBool(flags2.FlagAux)
	if isAux {
		return generateAuxSignerData(ctx, txf, msgs...)
	}

	// Only generate
	genOnly, _ := flagSet.GetBool(flags2.FlagGenerateOnly)
	if genOnly {
		return generateOnly(ctx, txf, msgs...)
	}

	// Simulate
	// Broadcast

	return GenerateOrBroadcastTxWithFactory(ctx, txf, msgs...)
}

func newFactory(ctx client.Context, flagSet *pflag.FlagSet, msgs ...transaction.Msg) (Factory, error) {
	k, err := keyring.NewAutoCLIKeyring(ctx.Keyring, ctx.AddressCodec)
	if err != nil {
		return Factory{}, err
	}

	params, err := txParamsFromFlagSet(flagSet, k)
	if err != nil {
		return Factory{}, err
	}

	txConfig, err := NewTxConfig(ConfigOptions{
		AddressCodec:          ctx.AddressCodec,
		Cdc:                   ctx.Codec,
		ValidatorAddressCodec: ctx.ValidatorAddressCodec,
	})
	if err != nil {
		return Factory{}, err
	}

	accRetriever := newAccountRetriever(ctx.AddressCodec, ctx, ctx.InterfaceRegistry)
	txf, err := NewFactory(k, ctx.Codec, accRetriever, txConfig, ctx.AddressCodec, ctx, params)
	if err != nil {
		return Factory{}, err
	}

	return txf, nil
}

// validateMessages validates all msgs before generating or broadcasting the tx.
// We were calling ValidateBasic separately in each CLI handler before.
// Right now, we're factorizing that call inside this function.
// ref: https://github.com/cosmos/cosmos-sdk/pull/9236#discussion_r623803504
func validateMessages(msgs ...transaction.Msg) error {
	for _, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic) // TODO: sdk dependency
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// generateAuxSignerData simply generates and prints the AuxSignerData.
func generateAuxSignerData(ctx client.Context, txf Factory, msgs ...transaction.Msg) error {
	auxSignerData, err := makeAuxSignerData(ctx, txf, msgs...)
	if err != nil {
		return err
	}

	return ctx.PrintString(auxSignerData.String())
}

func generateOnly(ctx client.Context, txf Factory, msgs ...transaction.Msg) error {
	uTx, err := txf.PrintUnsignedTx(msgs...)
	if err != nil {
		return err
	}
	return ctx.PrintString(uTx)
}

// GenerateOrBroadcastTxWithFactory will either generate and print an unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxWithFactory(clientCtx client.Context, txf Factory, msgs ...transaction.Msg) error {
	// Validate all msgs before generating or broadcasting the tx.
	// We were calling ValidateBasic separately in each CLI handler before.
	// Right now, we're factorizing that call inside this function.
	// ref: https://github.com/cosmos/cosmos-sdk/pull/9236#discussion_r623803504
	err := validateMessages(msgs...)
	if err != nil {
		return err
	}

	if clientCtx.GenerateOnly {
		uTx, err := txf.PrintUnsignedTx(msgs...)
		if err != nil {
			return err
		}
		return clientCtx.PrintString(uTx)
	}

	return BroadcastTx(clientCtx, txf, msgs...)
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func BroadcastTx(clientCtx client.Context, txf Factory, msgs ...transaction.Msg) error {
	txf, err := txf.Prepare()
	if err != nil {
		return err
	}

	if txf.SimulateAndExecute() || clientCtx.Simulate {
		if clientCtx.Offline {
			return errors.New("cannot estimate gas in offline mode")
		}

		_, adjusted, err := CalculateGas(clientCtx, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: txf.Gas()})
	}

	if clientCtx.Simulate {
		return nil
	}

	builder, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return err
	}

	if !clientCtx.SkipConfirm {
		encoder := txf.txConfig.TxJSONEncoder()
		if encoder == nil {
			return errors.New("failed to encode transaction: tx json encoder is nil")
		}

		unsigTx, err := builder.GetTx()
		if err != nil {
			return err
		}
		txBytes, err := encoder(unsigTx)
		if err != nil {
			return fmt.Errorf("failed to encode transaction: %w", err)
		}

		if err := clientCtx.PrintRaw(txBytes); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n%s\n", err, txBytes)
		}

		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stderr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\ncanceled transaction\n", err)
			return err
		}
		if !ok {
			_, _ = fmt.Fprintln(os.Stderr, "canceled transaction")
			return nil
		}
	}

	if err = txf.Sign(clientCtx.CmdContext, clientCtx.FromName, builder, true); err != nil {
		return err
	}

	signedTx, err := builder.GetTx()
	if err != nil {
		return err
	}

	txBytes, err := txf.txConfig.TxEncoder()(signedTx)
	if err != nil {
		return err
	}

	// broadcast to a CometBFT node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}

// CalculateGas simulates the execution of a transaction and returns the
// simulation response obtained by the query and the adjusted gas amount.
func CalculateGas(
	clientCtx gogogrpc.ClientConn, txf Factory, msgs ...transaction.Msg,
) (*tx.SimulateResponse, uint64, error) {
	txBytes, err := txf.BuildSimTx(msgs...)
	if err != nil {
		return nil, 0, err
	}

	txSvcClient := tx.NewServiceClient(clientCtx)
	simRes, err := txSvcClient.Simulate(context.Background(), &tx.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, 0, err
	}

	return simRes, uint64(txf.GasAdjustment() * float64(simRes.GasInfo.GasUsed)), nil
}

// makeAuxSignerData generates an AuxSignerData from the client inputs.
func makeAuxSignerData(clientCtx client.Context, f Factory, msgs ...transaction.Msg) (*apitx.AuxSignerData, error) {
	b := NewAuxTxBuilder()
	fromAddress, name, _, err := client.GetFromFields(clientCtx, clientCtx.Keyring, clientCtx.From)
	if err != nil {
		return nil, err
	}

	b.SetAddress(fromAddress.String())
	if f.txParams.offline {
		b.SetAccountNumber(f.AccountNumber())
		b.SetSequence(f.Sequence())
	} else {
		accNum, seq, err := f.accountRetriever.GetAccountNumberSequence(context.Background(), fromAddress)
		if err != nil {
			return nil, err
		}
		b.SetAccountNumber(accNum)
		b.SetSequence(seq)
	}

	err = b.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	err = b.SetSignMode(f.SignMode())
	if err != nil {
		return nil, err
	}

	pubKey, err := f.keybase.GetPubKey(name)
	if err != nil {
		return nil, err
	}

	err = b.SetPubKey(pubKey)
	if err != nil {
		return nil, err
	}

	b.SetChainID(clientCtx.ChainID)
	signBz, err := b.GetSignBytes()
	if err != nil {
		return nil, err
	}

	sig, err := f.keybase.Sign(name, signBz, f.SignMode())
	if err != nil {
		return nil, err
	}
	b.SetSignature(sig)

	return b.GetAuxSignerData()
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
			return sdkerrors.ErrNotSupported.Wrap("txs signed with CLI can have maximum 1 DIRECT signer")
		}
	}

	return nil
}

// countDirectSigners counts the number of DIRECT signers in a signature data.
func countDirectSigners(sigData SignatureData) int {
	switch data := sigData.(type) {
	case *SingleSignatureData:
		if data.SignMode == apitxsigning.SignMode_SIGN_MODE_DIRECT {
			return 1
		}

		return 0
	case *MultiSignatureData:
		directSigners := 0
		for _, d := range data.Signatures {
			directSigners += countDirectSigners(d)
		}

		return directSigners
	default:
		panic("unreachable case")
	}
}

func getSignMode(mode string) apitxsigning.SignMode {
	switch mode {
	case "direct":
		return apitxsigning.SignMode_SIGN_MODE_DIRECT
	case "direct-aux":
		return apitxsigning.SignMode_SIGN_MODE_DIRECT_AUX
	case "amino-json":
		return apitxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	case "textual":
		return apitxsigning.SignMode_SIGN_MODE_TEXTUAL
	}
	return apitxsigning.SignMode_SIGN_MODE_UNSPECIFIED
}
