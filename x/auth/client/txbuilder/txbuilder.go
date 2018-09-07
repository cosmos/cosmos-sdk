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

// TxBuilder implements a transaction context created in SDK modules.
type TxBuilder struct {
	Codec         *wire.Codec
	AccountNumber int64
	Sequence      int64
	Gas           int64
	ChainID       string
	Memo          string
	Fee           string
}

// NewTxBuilderFromCLI returns a new initialized TxBuilder with parameters from
// the command line using Viper.
func NewTxBuilderFromCLI() TxBuilder {
	// if chain ID is not specified manually, read default chain ID
	chainID := viper.GetString(client.FlagChainID)
	if chainID == "" {
		defaultChainID, err := defaultChainID()
		if err != nil {
			chainID = defaultChainID
		}
	}

	return TxBuilder{
		ChainID:       chainID,
		Gas:           viper.GetInt64(client.FlagGas),
		AccountNumber: viper.GetInt64(client.FlagAccountNumber),
		Sequence:      viper.GetInt64(client.FlagSequence),
		Fee:           viper.GetString(client.FlagFee),
		Memo:          viper.GetString(client.FlagMemo),
	}
}

// WithCodec returns a copy of the context with an updated codec.
func (bld TxBuilder) WithCodec(cdc *wire.Codec) TxBuilder {
	bld.Codec = cdc
	return bld
}

// WithChainID returns a copy of the context with an updated chainID.
func (bld TxBuilder) WithChainID(chainID string) TxBuilder {
	bld.ChainID = chainID
	return bld
}

// WithGas returns a copy of the context with an updated gas.
func (bld TxBuilder) WithGas(gas int64) TxBuilder {
	bld.Gas = gas
	return bld
}

// WithFee returns a copy of the context with an updated fee.
func (bld TxBuilder) WithFee(fee string) TxBuilder {
	bld.Fee = fee
	return bld
}

// WithSequence returns a copy of the context with an updated sequence number.
func (bld TxBuilder) WithSequence(sequence int64) TxBuilder {
	bld.Sequence = sequence
	return bld
}

// WithMemo returns a copy of the context with an updated memo.
func (bld TxBuilder) WithMemo(memo string) TxBuilder {
	bld.Memo = memo
	return bld
}

// WithAccountNumber returns a copy of the context with an account number.
func (bld TxBuilder) WithAccountNumber(accnum int64) TxBuilder {
	bld.AccountNumber = accnum
	return bld
}

// Build builds a single message to be signed from a TxBuilder given a set of
// messages. It returns an error if a fee is supplied but cannot be parsed.
func (bld TxBuilder) Build(msgs []sdk.Msg) (auth.StdSignMsg, error) {
	chainID := bld.ChainID
	if chainID == "" {
		return auth.StdSignMsg{}, errors.Errorf("chain ID required but not specified")
	}

	fee := sdk.Coin{}
	if bld.Fee != "" {
		parsedFee, err := sdk.ParseCoin(bld.Fee)
		if err != nil {
			return auth.StdSignMsg{}, err
		}

		fee = parsedFee
	}

	return auth.StdSignMsg{
		ChainID:       bld.ChainID,
		AccountNumber: bld.AccountNumber,
		Sequence:      bld.Sequence,
		Memo:          bld.Memo,
		Msgs:          msgs,
		Fee:           auth.NewStdFee(bld.Gas, fee),
	}, nil
}

// Sign signs a transaction given a name, passphrase, and a single message to
// signed. An error is returned if signing fails.
func (bld TxBuilder) Sign(name, passphrase string, msg auth.StdSignMsg) ([]byte, error) {
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

	return bld.Codec.MarshalBinary(auth.NewStdTx(msg.Msgs, msg.Fee, sigs, msg.Memo))
}

// BuildAndSign builds a single message to be signed, and signs a transaction
// with the built message given a name, passphrase, and a set of
// messages.
func (bld TxBuilder) BuildAndSign(name, passphrase string, msgs []sdk.Msg) ([]byte, error) {
	msg, err := bld.Build(msgs)
	if err != nil {
		return nil, err
	}

	return bld.Sign(name, passphrase, msg)
}

// BuildWithPubKey builds a single message to be signed from a TxBuilder given a set of
// messages and attach the public key associated to the given name.
// It returns an error if a fee is supplied but cannot be parsed or the key cannot be
// retrieved.
func (bld TxBuilder) BuildWithPubKey(name string, msgs []sdk.Msg) ([]byte, error) {
	msg, err := bld.Build(msgs)
	if err != nil {
		return nil, err
	}

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	info, err := keybase.Get(name)
	if err != nil {
		return nil, err
	}

	sigs := []auth.StdSignature{{
		AccountNumber: msg.AccountNumber,
		Sequence:      msg.Sequence,
		PubKey:        info.GetPubKey(),
	}}

	return bld.Codec.MarshalBinary(auth.NewStdTx(msg.Msgs, msg.Fee, sigs, msg.Memo))
}
