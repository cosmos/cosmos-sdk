package commands

import (
	"github.com/spf13/cobra"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"

	btypes "github.com/tendermint/basecoin/types"
)

var AccountQueryCmd = &cobra.Command{
	Use:   "account [address]",
	Short: "Get details of an account, with proof",
	RunE:  doAccountQuery,
}

func doAccountQuery(cmd *cobra.Command, args []string) error {
	height := proofcmd.GetHeight()
	addr, err := proofcmd.ParseHexKey(args, "address")
	if err != nil {
		return err
	}
	key := btypes.AccountKey(addr)

	// get the proof -> this will be used by all prover commands
	node := commands.GetNode()
	prover := proofs.NewAppProver(node)
	proof, err := proofcmd.GetProof(node, prover, key, height)
	if err != nil {
		return err
	}

	acc := new(btypes.Account)
	err = wire.ReadBinaryBytes(proof.Data(), &acc)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(acc, proof.BlockHeight())
}
