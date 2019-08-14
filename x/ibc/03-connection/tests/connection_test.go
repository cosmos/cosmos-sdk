package connection

import (
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
	node.OpenInit(t)
	header := node.Commit()

	// counterparty.OpenTry
	node.Counterparty.UpdateClient(t, header)
	cliobj := node.CLIObject()
	_, pconn := node.QueryValue(t, cliobj.Connection)
	_, pstate := node.QueryValue(t, cliobj.State)
	_, ptimeout := node.QueryValue(t, cliobj.NextTimeout)
	_, pcounterclient := node.QueryValue(t, cliobj.CounterpartyClient)
	// TODO: implement consensus state checking
	// _, pclient := node.Query(t, cliobj.Client.ConsensusStateKey)
	node.Counterparty.OpenTry(t, pconn, pstate, ptimeout, pcounterclient)
	header = node.Counterparty.Commit()

	// self.OpenAck
	node.UpdateClient(t, header)
	cliobj = node.Counterparty.CLIObject()
	_, pconn = node.Counterparty.QueryValue(t, cliobj.Connection)
	_, pstate = node.Counterparty.QueryValue(t, cliobj.State)
	_, ptimeout = node.Counterparty.QueryValue(t, cliobj.NextTimeout)
	_, pcounterclient = node.Counterparty.QueryValue(t, cliobj.CounterpartyClient)
	node.OpenAck(t, pconn, pstate, ptimeout, pcounterclient)
	header = node.Commit()

	// counterparty.OpenConfirm
	node.Counterparty.UpdateClient(t, header)
	cliobj = node.CLIObject()
	_, pstate = node.QueryValue(t, cliobj.State)
	_, ptimeout = node.QueryValue(t, cliobj.NextTimeout)
	node.Counterparty.OpenConfirm(t, pstate, ptimeout)
}
