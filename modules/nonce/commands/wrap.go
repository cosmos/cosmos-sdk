package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	"github.com/tendermint/basecoin/modules/nonce"
)

// nolint
const (
	FlagSequence = "sequence"
	FlagNonceKey = "nonce-key"
)

// NonceWrapper wraps a tx with a nonce
type NonceWrapper struct{}

var _ bcmd.Wrapper = NonceWrapper{}

// Wrap grabs the sequence number from the flag and wraps
// the tx with this nonce.  Grabs the permission from the signer,
// as we still only support single sig on the cli
func (NonceWrapper) Wrap(tx basecoin.Tx) (res basecoin.Tx, err error) {
	//add the nonce tx layer to the tx
	seq := viper.GetInt(FlagSequence)
	if seq < 0 {
		return res, fmt.Errorf("sequence must be greater than 0")
	}
	signers, err := readNonceKey()
	if err != nil {
		return res, err
	}
	res = nonce.NewTx(uint32(seq), signers, tx)
	return
}

// Register adds the sequence flags to the cli
func (NonceWrapper) Register(fs *pflag.FlagSet) {
	fs.Int(FlagSequence, -1, "Sequence number for this transaction")
	fs.String(FlagNonceKey, "", "Set of comma-separated addresses for the nonce (for multisig)")
}

func readNonceKey() (signers []basecoin.Actor, err error) {
	nonce := viper.GetString(FlagNonceKey)
	if nonce == "" {
		return []basecoin.Actor{bcmd.GetSignerAct()}, nil
	}

	var act basecoin.Actor
	for _, k := range strings.Split(nonce, ",") {
		act, err = bcmd.ParseAddress(k)
		if err != nil {
			return
		}
		signers = append(signers, act)
	}
	return
}
