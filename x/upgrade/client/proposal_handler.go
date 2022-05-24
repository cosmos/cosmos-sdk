package client

import (
	govclient "github.com/Stride-Labs/cosmos-sdk/x/gov/client"
	"github.com/Stride-Labs/cosmos-sdk/x/upgrade/client/cli"
	"github.com/Stride-Labs/cosmos-sdk/x/upgrade/client/rest"
)

var ProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitUpgradeProposal, rest.ProposalRESTHandler)
var CancelProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitCancelUpgradeProposal, rest.ProposalCancelRESTHandler)
