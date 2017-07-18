package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin"
	wire "github.com/tendermint/go-wire"
	lc "github.com/tendermint/light-client"
	lcmd "github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"

	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/stack"
)

// AccountQueryCmd - command to query an account
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
	key := stack.PrefixedKey(coin.NameCoin, auth.SigPerm(addr).Bytes())

	acc := coin.Account{}
	proof, err := proofcmd.GetAndParseAppProof(key, &acc)
	if lc.IsNoDataErr(err) {
		return errors.Errorf("Account bytes are empty for address %X ", addr)
	} else if err != nil {
		return err
	}

	return proofcmd.OutputProof(acc, proof.BlockHeight())
}

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
		nonce.NameNonce,
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

// BaseTxPresenter this decodes all basecoin tx
type BaseTxPresenter struct {
	proofs.RawPresenter // this handles MakeKey as hex bytes
}

// ParseData - unmarshal raw bytes to a basecoin tx
func (BaseTxPresenter) ParseData(raw []byte) (interface{}, error) {
	var tx basecoin.Tx
	err := wire.ReadBinaryBytes(raw, &tx)
	return tx, err
}
