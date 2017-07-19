package commands

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	lc "github.com/tendermint/light-client"

	"github.com/tendermint/basecoin"
	lcmd "github.com/tendermint/basecoin/client/commands"
	proofcmd "github.com/tendermint/basecoin/client/commands/proofs"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/stack"
)

// NonceQueryCmd - command to query an nonce account
var NonceQueryCmd = &cobra.Command{
	Use:   "nonce [address]",
	Short: "Get details of a nonce sequence number, with proof",
	RunE:  lcmd.RequireInit(nonceQueryCmd),
}

func nonceQueryCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("Missing required argument [address]")
	}
	addr := strings.Join(args, ",")

	signers, err := parseActors(addr)
	if err != nil {
		return err
	}

	seq, proof, err := doNonceQuery(signers)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(seq, proof.BlockHeight())
}

func doNonceQuery(signers []basecoin.Actor) (sequence uint32, proof lc.Proof, err error) {

	key := stack.PrefixedKey(nonce.NameNonce, nonce.GetSeqKey(signers))

	proof, err = proofcmd.GetAndParseAppProof(key, &sequence)
	if lc.IsNoDataErr(err) {
		// no data, return sequence 0
		return 0, proof, nil
	}

	return
}
