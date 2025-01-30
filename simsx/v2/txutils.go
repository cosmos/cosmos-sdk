package v2

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"google.golang.org/protobuf/types/known/anypb"

	"cosmossdk.io/core/transaction"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

const DefaultGenTxGas = 10_000_000

type Tx = transaction.Tx

// TXBuilder abstract transaction builder
type TXBuilder[T Tx] interface {
	// Build creates a signed transaction
	Build(ctx context.Context,
		ak simsx.AccountSource,
		senders []simsx.SimAccount,
		msg sdk.Msg,
		r *rand.Rand,
		chainID string,
	) (T, error)
}

var _ TXBuilder[Tx] = TXBuilderFn[Tx](nil)

// TXBuilderFn adapter that implements the TXBuilder interface
type TXBuilderFn[T Tx] func(ctx context.Context, ak simsx.AccountSource, senders []simsx.SimAccount, msg sdk.Msg, r *rand.Rand, chainID string) (T, error)

func (b TXBuilderFn[T]) Build(ctx context.Context, ak simsx.AccountSource, senders []simsx.SimAccount, msg sdk.Msg, r *rand.Rand, chainID string) (T, error) {
	return b(ctx, ak, senders, msg, r, chainID)
}

// NewSDKTXBuilder constructor to create a signed transaction builder for sdk.Tx type.
func NewSDKTXBuilder[T Tx](txConfig client.TxConfig, defaultGas uint64) TXBuilder[T] {
	return TXBuilderFn[T](func(ctx context.Context, ak simsx.AccountSource, senders []simsx.SimAccount, msg sdk.Msg, r *rand.Rand, chainID string) (tx T, err error) {
		accountNumbers := make([]uint64, len(senders))
		sequenceNumbers := make([]uint64, len(senders))
		for i := 0; i < len(senders); i++ {
			acc := ak.GetAccount(ctx, senders[i].Address)
			accountNumbers[i] = acc.GetAccountNumber()
			sequenceNumbers[i] = acc.GetSequence()
		}
		fees := senders[0].LiquidBalance().RandFees()
		sdkTx, err := GenSignedMockTx(
			r,
			txConfig,
			[]sdk.Msg{msg},
			fees,
			defaultGas,
			chainID,
			accountNumbers,
			sequenceNumbers,
			simsx.Collect(senders, func(a simsx.SimAccount) cryptotypes.PrivKey { return a.PrivKey })...,
		)
		if err != nil {
			return tx, err
		}
		out, ok := sdkTx.(T)
		if !ok {
			return out, errors.New("unexpected Tx type")
		}
		return out, nil
	})
}

// GenSignedMockTx generates a signed mock transaction.
func GenSignedMockTx(
	r *rand.Rand,
	txConfig client.TxConfig,
	msgs []sdk.Msg,
	feeAmt sdk.Coins,
	gas uint64,
	chainID string,
	accNums, accSeqs []uint64,
	priv ...cryptotypes.PrivKey,
) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))

	// create a random length memo
	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: txConfig.SignModeHandler().DefaultMode(),
			},
			Sequence: accSeqs[i],
		}
	}

	tx := txConfig.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		anyPk, err := codectypes.NewAnyWithValue(p.PubKey())
		if err != nil {
			return nil, err
		}

		signerData := txsigning.SignerData{
			Address:       sdk.AccAddress(p.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        &anypb.Any{TypeUrl: anyPk.TypeUrl, Value: anyPk.Value},
		}

		signBytes, err := authsign.GetSignBytesAdapter(
			context.Background(), txConfig.SignModeHandler(), txConfig.SignModeHandler().DefaultMode(), signerData,
			tx.GetTx())
		if err != nil {
			return nil, fmt.Errorf("sign bytes: %w", err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			return nil, fmt.Errorf("sign: %w", err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
	}

	if err = tx.SetSignatures(sigs...); err != nil {
		return nil, fmt.Errorf("signature: %w", err)
	}

	return tx.GetTx(), nil
}

var _ transaction.Codec[Tx] = &GenericTxDecoder[Tx]{}

// GenericTxDecoder Encoder type that implements transaction.Codec
type GenericTxDecoder[T Tx] struct {
	txConfig client.TxConfig
}

// NewGenericTxDecoder constructor
func NewGenericTxDecoder[T Tx](txConfig client.TxConfig) *GenericTxDecoder[T] {
	return &GenericTxDecoder[T]{txConfig: txConfig}
}

// Decode implements transaction.Codec.
func (t *GenericTxDecoder[T]) Decode(bz []byte) (T, error) {
	var out T
	tx, err := t.txConfig.TxDecoder()(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(T)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}

// DecodeJSON implements transaction.Codec.
func (t *GenericTxDecoder[T]) DecodeJSON(bz []byte) (T, error) {
	var out T
	tx, err := t.txConfig.TxJSONDecoder()(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(T)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}
