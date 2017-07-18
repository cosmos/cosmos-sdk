package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	lc "github.com/tendermint/light-client"
	lcmd "github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/stack"
)

// NonceQueryCmd - command to query an nonce account
var NonceQueryCmd = &cobra.Command{
	Use:   "nonce [address]",
	Short: "Get details of a nonce sequence number, with proof",
	RunE:  lcmd.RequireInit(doNonceQuery),
}

func doNonceQuery(cmd *cobra.Command, args []string) error {
	addr, err := proofcmd.ParseHexKey(args, "address")
	if err != nil {
		return err
	}

	act := []basecoin.Actor{basecoin.NewActor(
		auth.NameSigs,
		addr,
	)}

	key := stack.PrefixedKey(nonce.NameNonce, nonce.GetSeqKey(act))

	var seq uint32
	proof, err := proofcmd.GetAndParseAppProof(key, &seq)
	if lc.IsNoDataErr(err) {
		return errors.Errorf("Sequence is empty for address %X ", addr)
	} else if err != nil {
		return err
	}

	return proofcmd.OutputProof(seq, proof.BlockHeight())
}
