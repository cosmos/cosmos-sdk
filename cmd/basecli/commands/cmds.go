package commands

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	cmn "github.com/tendermint/tmlibs/common"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/auth"
)

var (
	// Middleware must be set in main.go to defined the wrappers we should apply
	Middleware Wrapper
)

// Wrapper defines the information needed for each middleware package that
// wraps the data.  They should read all configuration out of bounds via viper.
type Wrapper interface {
	Wrap(basecoin.Tx) (basecoin.Tx, error)
	Register(*pflag.FlagSet)
}

// Wrappers combines a list of wrapper middlewares.
// The first one is the inner-most layer, eg. Fee, Nonce, Chain, Auth
type Wrappers []Wrapper

var _ Wrapper = Wrappers{}

// Wrap applies the wrappers to the passed in tx in order,
// aborting on the first error
func (ws Wrappers) Wrap(tx basecoin.Tx) (basecoin.Tx, error) {
	var err error
	for _, w := range ws {
		tx, err = w.Wrap(tx)
		if err != nil {
			break
		}
	}
	return tx, err
}

// Register adds any needed flags to the command
func (ws Wrappers) Register(fs *pflag.FlagSet) {
	for _, w := range ws {
		w.Register(fs)
	}
}

// ValidateResult returns an appropriate error if the server rejected the
// tx in CheckTx or DeliverTx
func ValidateResult(res *ctypes.ResultBroadcastTxCommit) error {
	if res.CheckTx.IsErr() {
		return fmt.Errorf("CheckTx: (%d): %s", res.CheckTx.Code, res.CheckTx.Log)
	}
	if res.DeliverTx.IsErr() {
		return fmt.Errorf("DeliverTx: (%d): %s", res.DeliverTx.Code, res.DeliverTx.Log)
	}
	return nil
}

// ParseAddress parses an address of form:
// [<chain>:][<app>:]<hex address>
// into a basecoin.Actor.
// If app is not specified or "", then assume auth.NameSigs
func ParseAddress(input string) (res basecoin.Actor, err error) {
	chain, app := "", auth.NameSigs
	input = strings.TrimSpace(input)
	spl := strings.SplitN(input, ":", 3)

	if len(spl) == 3 {
		chain = spl[0]
		spl = spl[1:]
	}
	if len(spl) == 2 {
		if spl[0] != "" {
			app = spl[0]
		}
		spl = spl[1:]
	}

	addr, err := hex.DecodeString(cmn.StripHex(spl[0]))
	if err != nil {
		return res, errors.Errorf("Address is invalid hex: %v\n", err)
	}
	res = basecoin.Actor{
		ChainID: chain,
		App:     app,
		Address: addr,
	}
	return
}

// GetSignerAct returns the address of the signer of the tx
// (as we still only support single sig)
func GetSignerAct() (res basecoin.Actor) {
	// this could be much cooler with multisig...
	signer := txcmd.GetSigner()
	if !signer.Empty() {
		res = auth.SigPerm(signer.Address())
	}
	return res
}
