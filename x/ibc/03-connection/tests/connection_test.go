package connection

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	tmclient "github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint/tests"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func registerCodec(cdc *codec.Codec) {
	client.RegisterCodec(cdc)
	tmclient.RegisterCodec(cdc)
	commitment.RegisterCodec(cdc)
	merkle.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
}

func TestHandshake(t *testing.T) {
	cdc := codec.New()
	registerCodec(cdc)

	node := NewNode(tendermint.NewMockValidators(100, 10), tendermint.NewMockValidators(100, 10), cdc)
	node.Commit()
	node.Counterparty.Commit()

	node.CreateClient(t)
	node.Counterparty.CreateClient(t)

	// self.OpenInit
	node.Advance(t)
	header := node.Commit()

	fmt.Printf("ddd\n\n\n\n%+v\n", node.Root)

	// counterparty.OpenTry
	node.Counterparty.UpdateClient(t, header)
	cliobj := node.CLIObject()
	_, pconn := node.Query(t, cliobj.ConnectionKey)
	fmt.Printf("\n\n\n\n%s: %+v, %s\n", cliobj.ConnectionKey, pconn.(merkle.Proof).Proof, string(pconn.(merkle.Proof).Key))
	_, pstate := node.Query(t, cliobj.StateKey)
	_, ptimeout := node.Query(t, cliobj.NextTimeoutKey)
	// TODO: implement consensus state checking
	// _, pclient := node.Query(t, cliobj.Client.ConsensusStateKey)
	node.Advance(t, pconn, pstate, ptimeout)

	// self.OpenAck
	node.Advance(t)

	// counterparty.OpenConfirm
	node.Advance(t)
}
