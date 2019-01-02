package context

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// TxBuilder implements a transaction context created in SDK modules.
type TxBuilder struct {
	txEncoder          sdk.TxEncoder
	accountNumber      uint64
	sequence           uint64
	gas                uint64
	gasAdjustment      float64
	simulateAndExecute bool
	chainID            string
	memo               string
	fees               sdk.Coins
}

// NewTxBuilder returns a new initialized TxBuilder
func NewTxBuilder(txEncoder sdk.TxEncoder, accNumber, seq, gas uint64, gasAdj float64, simulateAndExecute bool, chainID, memo string, fees sdk.Coins) TxBuilder {
	return TxBuilder{
		txEncoder:          txEncoder,
		accountNumber:      accNumber,
		sequence:           seq,
		gas:                gas,
		gasAdjustment:      gasAdj,
		simulateAndExecute: simulateAndExecute,
		chainID:            chainID,
		memo:               memo,
		fees:               fees,
	}
}

// NewTxBuilderFromCLI returns a new initialized TxBuilder with parameters from
// the command line using Viper.
func NewTxBuilderFromCLI() TxBuilder {
	txbldr := TxBuilder{
		accountNumber:      uint64(viper.GetInt64(client.FlagAccountNumber)),
		sequence:           uint64(viper.GetInt64(client.FlagSequence)),
		gas:                client.GasFlagVar.Gas,
		gasAdjustment:      viper.GetFloat64(client.FlagGasAdjustment),
		simulateAndExecute: client.GasFlagVar.Simulate,
		chainID:            viper.GetString(client.FlagChainID),
		memo:               viper.GetString(client.FlagMemo),
	}
	return txbldr.WithFees(viper.GetString(client.FlagFees))
}

// GetTxEncoder returns the transaction encoder
func (bldr TxBuilder) GetTxEncoder() sdk.TxEncoder { return bldr.txEncoder }

// GetAccountNumber returns the account number
func (bldr TxBuilder) GetAccountNumber() uint64 { return bldr.accountNumber }

// GetSequence returns the transaction sequence
func (bldr TxBuilder) GetSequence() uint64 { return bldr.sequence }

// GetGas returns the gas for the transaction
func (bldr TxBuilder) GetGas() uint64 { return bldr.gas }

// GetGasAdjustment returns the gas adjustment
func (bldr TxBuilder) GetGasAdjustment() float64 { return bldr.gasAdjustment }

// GetSimulateAndExecute returns the option to simulate and then execute the transaction
// using the gas from the simulation results
func (bldr TxBuilder) GetSimulateAndExecute() bool { return bldr.simulateAndExecute }

// GetChainID returns the chain id
func (bldr TxBuilder) GetChainID() string { return bldr.chainID }

// GetMemo returns the memo message
func (bldr TxBuilder) GetMemo() string { return bldr.memo }

// GetFees returns the fees for the transaction
func (bldr TxBuilder) GetFees() sdk.Coins { return bldr.fees }

// WithTxEncoder returns a copy of the context with an updated codec.
func (bldr TxBuilder) WithTxEncoder(txEncoder sdk.TxEncoder) TxBuilder {
	bldr.txEncoder = txEncoder
	return bldr
}

// WithChainID returns a copy of the context with an updated chainID.
func (bldr TxBuilder) WithChainID(chainID string) TxBuilder {
	bldr.chainID = chainID
	return bldr
}

// WithGas returns a copy of the context with an updated gas.
func (bldr TxBuilder) WithGas(gas uint64) TxBuilder {
	bldr.gas = gas
	return bldr
}

// WithFees returns a copy of the context with an updated fee.
func (bldr TxBuilder) WithFees(fees string) TxBuilder {
	parsedFees, err := sdk.ParseCoins(fees)
	if err != nil {
		panic(err)
	}
	bldr.fees = parsedFees
	return bldr
}

// WithSequence returns a copy of the context with an updated sequence number.
func (bldr TxBuilder) WithSequence(sequence uint64) TxBuilder {
	bldr.sequence = sequence
	return bldr
}

// WithMemo returns a copy of the context with an updated memo.
func (bldr TxBuilder) WithMemo(memo string) TxBuilder {
	bldr.memo = strings.TrimSpace(memo)
	return bldr
}

// WithAccountNumber returns a copy of the context with an account number.
func (bldr TxBuilder) WithAccountNumber(accnum uint64) TxBuilder {
	bldr.accountNumber = accnum
	return bldr
}

// Build builds a single message to be signed from a TxBuilder given a set of
// messages. It returns an error if a fee is supplied but cannot be parsed.
func (bldr TxBuilder) Build(msgs []sdk.Msg) (StdSignMsg, error) {
	chainID := bldr.chainID
	if chainID == "" {
		return StdSignMsg{}, errors.Errorf("chain ID required but not specified")
	}

	return StdSignMsg{
		ChainID:       bldr.chainID,
		AccountNumber: bldr.accountNumber,
		Sequence:      bldr.sequence,
		Memo:          bldr.memo,
		Msgs:          msgs,
		Fee:           auth.NewStdFee(bldr.gas, bldr.fees),
	}, nil
}

// Sign signs a transaction given a name, passphrase, and a single message to
// signed. An error is returned if signing fails.
func (bldr TxBuilder) Sign(name, passphrase string, msg StdSignMsg) ([]byte, error) {
	sig, err := MakeSignature(name, passphrase, msg)
	if err != nil {
		return nil, err
	}

	return bldr.txEncoder(auth.NewStdTx(msg.Msgs, msg.Fee, []auth.StdSignature{sig}, msg.Memo))
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
		PubKey: info.GetPubKey(),
	}}

	return bldr.txEncoder(auth.NewStdTx(msg.Msgs, msg.Fee, sigs, msg.Memo))
}

// SignStdTx appends a signature to a StdTx and returns a copy of a it. If append
// is false, it replaces the signatures already attached with the new signature.
func (bldr TxBuilder) SignStdTx(name, passphrase string, stdTx auth.StdTx, appendSig bool) (signedStdTx auth.StdTx, err error) {
	stdSignature, err := MakeSignature(name, passphrase, StdSignMsg{
		ChainID:       bldr.chainID,
		AccountNumber: bldr.accountNumber,
		Sequence:      bldr.sequence,
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
		PubKey:    pubkey,
		Signature: sigBytes,
	}, nil
}
