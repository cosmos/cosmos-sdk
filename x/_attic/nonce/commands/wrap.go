package commands

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
)

// nolint
const (
	FlagSequence = "sequence"
	FlagNonceKey = "nonce-key"
)

// NonceWrapper wraps a tx with a nonce
type NonceWrapper struct{}

var _ txcmd.Wrapper = NonceWrapper{}

// Wrap grabs the sequence number from the flag and wraps
// the tx with this nonce.  Grabs the permission from the signer,
// as we still only support single sig on the cli
func (NonceWrapper) Wrap(tx sdk.Tx) (res sdk.Tx, err error) {

	signers, err := readNonceKey()
	if err != nil {
		return res, err
	}

	seq, err := readSequence(signers)
	if err != nil {
		return res, err
	}

	res = nonce.NewTx(seq, signers, tx)
	return
}

// Register adds the sequence flags to the cli
func (NonceWrapper) Register(fs *pflag.FlagSet) {
	fs.Int(FlagSequence, -1, "Sequence number for this transaction")
	fs.String(FlagNonceKey, "", "Set of comma-separated addresses for the nonce (for multisig)")
}

func readNonceKey() ([]sdk.Actor, error) {
	nonce := viper.GetString(FlagNonceKey)
	if nonce == "" {
		return []sdk.Actor{txcmd.GetSignerAct()}, nil
	}
	return commands.ParseActors(nonce)
}

// read the sequence from the flag or query for it if flag is -1
func readSequence(signers []sdk.Actor) (seq uint32, err error) {
	//add the nonce tx layer to the tx
	seqFlag := viper.GetInt(FlagSequence)

	switch {
	case seqFlag > 0:
		seq = uint32(seqFlag)

	case seqFlag == -1:
		//autocalculation for default sequence
		seq, _, err = doNonceQuery(signers)
		if err != nil {
			return
		}

		//increase the sequence by 1!
		seq++

	default:
		err = fmt.Errorf("sequence must be either greater than 0, or -1 for autocalculation")
	}

	return
}
