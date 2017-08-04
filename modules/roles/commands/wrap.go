package commands

import (
	"encoding/hex"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/basecoin"
	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/roles"
)

// nolint
const (
	FlagAssumeRole = "assume-role"
)

// RoleWrapper wraps a tx with 0, 1, or more roles
type RoleWrapper struct{}

var _ txcmd.Wrapper = RoleWrapper{}

// Wrap grabs the sequence number from the flag and wraps
// the tx with this nonce.  Grabs the permission from the signer,
// as we still only support single sig on the cli
func (RoleWrapper) Wrap(tx basecoin.Tx) (basecoin.Tx, error) {
	assume := viper.GetStringSlice(FlagAssumeRole)

	// we wrap from inside-out, so we must wrap them in the reverse order,
	// so they are applied in the order the user intended
	for i := len(assume) - 1; i >= 0; i-- {
		r, err := parseRole(assume[i])
		if err != nil {
			return tx, err
		}
		tx = roles.NewAssumeRoleTx(r, tx)
	}
	return tx, nil
}

// Register adds the sequence flags to the cli
func (RoleWrapper) Register(fs *pflag.FlagSet) {
	fs.StringSlice(FlagAssumeRole, nil, "Roles to assume (can use multiple times)")
}

// parse role turns the string->byte... todo: support hex?
func parseRole(role string) ([]byte, error) {
	res, err := hex.DecodeString(cmn.StripHex(role))
	if err != nil {
		err = errors.WithMessage("Address is invalid hex", err,
			abci.CodeType_EncodingError)
	}
	return res, err
}
