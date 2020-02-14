package client

import (
	"github.com/cosmos/cosmos-sdk/params/client/cli"
	"github.com/cosmos/cosmos-sdk/params/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// param change proposal handler
var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
