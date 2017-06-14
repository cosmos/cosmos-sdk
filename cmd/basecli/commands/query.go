package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"

	bccmd "github.com/tendermint/basecoin/cmd/commands"
	btypes "github.com/tendermint/basecoin/types"
)

func init() {
	//first modify the full node account query command for the light client
	bccmd.AccountCmd.RunE = accountCmd
	proofcmd.RootCmd.AddCommand(bccmd.AccountCmd)
}

func accountCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("account command requires an argument ([address])") //never stack trace
	}

	addrHex := StripHex(args[0])

	// convert destination address to bytes
	addr, err := hex.DecodeString(addrHex)
	if err != nil {
		return errors.Errorf("Account address (%v) is invalid hex: %v\n", addrHex, err)
	}

	// get the proof -> this will be used by all prover commands
	height := proofcmd.GetHeight()
	node := commands.GetNode()
	prover := proofs.NewAppProver(node)
	key := btypes.AccountKey(addr)
	proof, err := proofcmd.GetProof(node, prover, key, height)
	if err != nil {
		return err
	}

	var acc *btypes.Account
	err = wire.ReadBinaryBytes(proof.Data(), &acc)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(&acc, proof.BlockHeight())
}
