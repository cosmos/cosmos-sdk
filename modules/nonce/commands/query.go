package commands

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	lc "github.com/tendermint/light-client"
	lcmd "github.com/tendermint/basecoin/commands"
	proofcmd "github.com/tendermint/basecoin/commands/proofs"

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
	if len(args) == 0 {
		return errors.New("Missing required argument [address]")
	}
	addr := strings.Join(args, ",")
	act, err := parseActors(addr)
	if err != nil {
		return err
	}

	key := stack.PrefixedKey(nonce.NameNonce, nonce.GetSeqKey(act))

	var seq uint32
	proof, err := proofcmd.GetAndParseAppProof(key, &seq)
	if lc.IsNoDataErr(err) {
		return errors.Errorf("Sequence is empty for address %s ", addr)
	} else if err != nil {
		return err
	}

	return proofcmd.OutputProof(seq, proof.BlockHeight())
}
