package context

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// TxContext implements a transaction context created in SDK modules.
type TxContext struct {
	Codec         *wire.Codec
	AccountNumber int64
	Sequence      int64
	Gas           int64
	ChainID       string
	Memo          string
	Fee           string
}

// NewTxContextFromCLI returns a new initialized TxContext with parameters from
// the command line using Viper.
func NewTxContextFromCLI() TxContext {
	// if chain ID is not specified manually, read default chain ID
	chainID := viper.GetString(client.FlagChainID)
	if chainID == "" {
		defaultChainID, err := defaultChainID()
		if err != nil {
			chainID = defaultChainID
		}
	}

	return TxContext{
		ChainID:       chainID,
		Gas:           viper.GetInt64(client.FlagGas),
		AccountNumber: viper.GetInt64(client.FlagAccountNumber),
		Sequence:      viper.GetInt64(client.FlagSequence),
		Fee:           viper.GetString(client.FlagFee),
		Memo:          viper.GetString(client.FlagMemo),
	}
}

// WithCodec returns a copy of the context with an updated codec.
func (ctx TxContext) WithCodec(cdc *wire.Codec) TxContext {
	ctx.Codec = cdc
	return ctx
}

// WithChainID returns a copy of the context with an updated chainID.
func (ctx TxContext) WithChainID(chainID string) TxContext {
	ctx.ChainID = chainID
	return ctx
}

// WithGas returns a copy of the context with an updated gas.
func (ctx TxContext) WithGas(gas int64) TxContext {
	ctx.Gas = gas
	return ctx
}

// WithFee returns a copy of the context with an updated fee.
func (ctx TxContext) WithFee(fee string) TxContext {
	ctx.Fee = fee
	return ctx
}

// WithSequence returns a copy of the context with an updated sequence number.
func (ctx TxContext) WithSequence(sequence int64) TxContext {
	ctx.Sequence = sequence
	return ctx
}

// WithMemo returns a copy of the context with an updated memo.
func (ctx TxContext) WithMemo(memo string) TxContext {
	ctx.Memo = memo
	return ctx
}

// WithAccountNumber returns a copy of the context with an account number.
func (ctx TxContext) WithAccountNumber(accnum int64) TxContext {
	ctx.AccountNumber = accnum
	return ctx
}

// Build builds a single message to be signed from a TxContext given a set of
// messages. It returns an error if a fee is supplied but cannot be parsed.
func (ctx TxContext) Build(msgs []sdk.Msg) (auth.StdSignMsg, error) {
	chainID := ctx.ChainID
	if chainID == "" {
		return auth.StdSignMsg{}, errors.Errorf("chain ID required but not specified")
	}

	fee := sdk.Coin{}
	if ctx.Fee != "" {
		parsedFee, err := sdk.ParseCoin(ctx.Fee)
		if err != nil {
			return auth.StdSignMsg{}, err
		}

		fee = parsedFee
	}

	return auth.StdSignMsg{
		ChainID:       ctx.ChainID,
		AccountNumber: ctx.AccountNumber,
		Sequence:      ctx.Sequence,
		Memo:          ctx.Memo,
		Msgs:          msgs,

		// TODO: run simulate to estimate gas?
		Fee: auth.NewStdFee(ctx.Gas, fee),
	}, nil
}

// Sign signs a transaction given a name, passphrase, and a single message to
// signed. An error is returned if signing fails.
func (ctx TxContext) Sign(name, passphrase string, msg auth.StdSignMsg) ([]byte, error) {
	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	sig, pubkey, err := keybase.Sign(name, passphrase, msg.Bytes())
	if err != nil {
		return nil, err
	}

	sigs := []auth.StdSignature{{
		AccountNumber: msg.AccountNumber,
		Sequence:      msg.Sequence,
		PubKey:        pubkey,
		Signature:     sig,
	}}

	return ctx.Codec.MarshalBinary(auth.NewStdTx(msg.Msgs, msg.Fee, sigs, msg.Memo))
}

// BuildAndSign builds a single message to be signed, and signs a transaction
// with the built message given a name, passphrase, and a set of
// messages.
func (ctx TxContext) BuildAndSign(name, passphrase string, msgs []sdk.Msg) ([]byte, error) {
	msg, err := ctx.Build(msgs)
	if err != nil {
		return nil, err
	}

	return ctx.Sign(name, passphrase, msg)
}
