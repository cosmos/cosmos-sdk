package context

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// TxBuilder implements a transaction context created in SDK modules.
type TxBuilder struct {
	Codec         *codec.Codec
	AccountNumber uint64
	Sequence      uint64
	Gas           uint64
	GasAdjustment float64
	SimulateGas   bool
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
		defaultChainID, err := sdk.DefaultChainID()
		if err != nil {
			chainID = defaultChainID
		}
	}

	return TxBuilder{
		ChainID:       chainID,
		AccountNumber: uint64(viper.GetInt64(client.FlagAccountNumber)),
		Gas:           client.GasFlagVar.Gas,
		GasAdjustment: viper.GetFloat64(client.FlagGasAdjustment),
		Sequence:      uint64(viper.GetInt64(client.FlagSequence)),
		SimulateGas:   client.GasFlagVar.Simulate,
		Fee:           viper.GetString(client.FlagFee),
		Memo:          viper.GetString(client.FlagMemo),
	}
}

// WithCodec returns a copy of the context with an updated codec.
func (bldr TxBuilder) WithCodec(cdc *codec.Codec) TxBuilder {
	bldr.Codec = cdc
	return bldr
}

// WithChainID returns a copy of the context with an updated chainID.
func (bldr TxBuilder) WithChainID(chainID string) TxBuilder {
	bldr.ChainID = chainID
	return bldr
}

// WithGas returns a copy of the context with an updated gas.
func (bldr TxBuilder) WithGas(gas uint64) TxBuilder {
	bldr.Gas = gas
	return bldr
}

// WithFee returns a copy of the context with an updated fee.
func (bldr TxBuilder) WithFee(fee string) TxBuilder {
	bldr.Fee = fee
	return bldr
}

// WithSequence returns a copy of the context with an updated sequence number.
func (bldr TxBuilder) WithSequence(sequence uint64) TxBuilder {
	bldr.Sequence = sequence
	return bldr
}

// WithMemo returns a copy of the context with an updated memo.
func (bldr TxBuilder) WithMemo(memo string) TxBuilder {
	bldr.Memo = memo
	return bldr
}

// WithAccountNumber returns a copy of the context with an account number.
func (bldr TxBuilder) WithAccountNumber(accnum uint64) TxBuilder {
	bldr.AccountNumber = accnum
	return bldr
}

// Build builds a single message to be signed from a TxBuilder given a set of
// messages. It returns an error if a fee is supplied but cannot be parsed.
func (bldr TxBuilder) Build(msgs []sdk.Msg) (StdSignMsg, error) {
	chainID := bldr.ChainID
	if chainID == "" {
		return StdSignMsg{}, errors.Errorf("chain ID required but not specified")
	}

	fee := sdk.Coin{}
	if bldr.Fee != "" {
		parsedFee, err := sdk.ParseCoin(bldr.Fee)
		if err != nil {
			return StdSignMsg{}, err
		}

		fee = parsedFee
	}

	return StdSignMsg{
		ChainID:       bldr.ChainID,
		AccountNumber: bldr.AccountNumber,
		Sequence:      bldr.Sequence,
		Memo:          bldr.Memo,
		Msgs:          msgs,
		Fee:           auth.NewStdFee(bldr.Gas, fee),
	}, nil
}

// Sign signs a transaction given a name, passphrase, and a single message to
// signed. An error is returned if signing fails.
func (bldr TxBuilder) Sign(name, passphrase string, msg StdSignMsg) ([]byte, error) {
	sig, err := MakeSignature(name, passphrase, msg)
	if err != nil {
		return nil, err
	}

	return bldr.Codec.MarshalBinaryLengthPrefixed(auth.NewStdTx(msg.Msgs, msg.Fee, []auth.StdSignature{sig}, msg.Memo))
}

// BuildAndSign builds a single message to be signed, and signs a transaction
// with the built message given a name, passphrase, and a set of
// messages.
func (bldr TxBuilder) BuildAndSign(name, passphrase string, msgs []sdk.Msg) ([]byte, error) {
	msg, err := bldr.Build(msgs)
	if err != nil {
		return nil, err
	}

	return bldr.Sign(name, passphrase, msg)
}

// BuildWithPubKey builds a single message to be signed from a TxBuilder given a set of
// messages and attach the public key associated to the given name.
// It returns an error if a fee is supplied but cannot be parsed or the key cannot be
// retrieved.
func (bldr TxBuilder) BuildWithPubKey(name string, msgs []sdk.Msg) ([]byte, error) {
	msg, err := bldr.Build(msgs)
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

	return bldr.Codec.MarshalBinaryLengthPrefixed(auth.NewStdTx(msg.Msgs, msg.Fee, sigs, msg.Memo))
}

// SignStdTx appends a signature to a StdTx and returns a copy of a it. If append
// is false, it replaces the signatures already attached with the new signature.
func (bldr TxBuilder) SignStdTx(name, passphrase string, stdTx auth.StdTx, appendSig bool) (signedStdTx auth.StdTx, err error) {
	stdSignature, err := MakeSignature(name, passphrase, StdSignMsg{
		ChainID:       bldr.ChainID,
		AccountNumber: bldr.AccountNumber,
		Sequence:      bldr.Sequence,
		Fee:           stdTx.Fee,
		Msgs:          stdTx.GetMsgs(),
		Memo:          stdTx.GetMemo(),
	})
	if err != nil {
		return
	}

	sigs := stdTx.GetSignatures()
	if len(sigs) == 0 || !appendSig {
		sigs = []auth.StdSignature{stdSignature}
	} else {
		sigs = append(sigs, stdSignature)
	}
	signedStdTx = auth.NewStdTx(stdTx.GetMsgs(), stdTx.Fee, sigs, stdTx.GetMemo())
	return
}

// MakeSignature builds a StdSignature given key name, passphrase, and a StdSignMsg.
func MakeSignature(name, passphrase string, msg StdSignMsg) (sig auth.StdSignature, err error) {
	keybase, err := keys.GetKeyBase()
	if err != nil {
		return
	}
	sigBytes, pubkey, err := keybase.Sign(name, passphrase, msg.Bytes())
	if err != nil {
		return
	}
	return auth.StdSignature{
		AccountNumber: msg.AccountNumber,
		Sequence:      msg.Sequence,
		PubKey:        pubkey,
		Signature:     sigBytes,
	}, nil
}
