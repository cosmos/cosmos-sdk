package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	wire "github.com/tendermint/go-wire"
	lc "github.com/tendermint/light-client"
	lcmd "github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"

	btypes "github.com/tendermint/basecoin/types"
)

var AccountQueryCmd = &cobra.Command{
	Use:   "account [address]",
	Short: "Get details of an account, with proof",
	RunE:  lcmd.RequireInit(doAccountQuery),
}

func doAccountQuery(cmd *cobra.Command, args []string) error {
	addr, err := proofcmd.ParseHexKey(args, "address")
	if err != nil {
		return err
	}
	key := btypes.AccountKey(addr)

	acc := new(btypes.Account)
	proof, err := proofcmd.GetAndParseAppProof(key, &acc)
	if lc.IsNoDataErr(err) {
		return errors.Errorf("Account bytes are empty for address %X ", addr)
	} else if err != nil {
		return err
	}

	return proofcmd.OutputProof(acc, proof.BlockHeight())
}

// BaseTxPresenter this decodes all basecoin tx
type BaseTxPresenter struct {
	proofs.RawPresenter // this handles MakeKey as hex bytes
}

func (_ BaseTxPresenter) ParseData(raw []byte) (interface{}, error) {
	var tx btypes.TxS
	err := wire.ReadBinaryBytes(raw, &tx)
	return tx, err
}
