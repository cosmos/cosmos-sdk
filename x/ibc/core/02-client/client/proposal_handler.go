package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitUpdateClientProposal, nil)
