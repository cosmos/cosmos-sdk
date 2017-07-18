package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/client/commands"
	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	"github.com/tendermint/basecoin/modules/nonce"
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
func (NonceWrapper) Wrap(tx basecoin.Tx) (res basecoin.Tx, err error) {
	seq, err := readSequence()
	if err != nil {
		return res, err
	}

	signers, err := readNonceKey()
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

func readNonceKey() ([]basecoin.Actor, error) {
	nonce := viper.GetString(FlagNonceKey)
	if nonce == "" {
		return []basecoin.Actor{txcmd.GetSignerAct()}, nil
	}
	return parseActors(nonce)
}

func parseActors(key string) (signers []basecoin.Actor, err error) {
	var act basecoin.Actor
	for _, k := range strings.Split(key, ",") {
		act, err = commands.ParseAddress(k)
		if err != nil {
			return
		}
		signers = append(signers, act)
	}
	return
}

func readSequence() (uint32, error) {
	//add the nonce tx layer to the tx
	seq := viper.GetInt(FlagSequence)
	if seq > 0 {
		return uint32(seq), nil
	}

	// TODO: try to download from query..
	return 0, fmt.Errorf("sequence must be greater than 0")
}
