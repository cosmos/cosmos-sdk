package client

import (
	govclient "github.com/pointnetwork/cosmos-point-sdk/x/gov/client"
	"github.com/pointnetwork/cosmos-point-sdk/x/upgrade/client/cli"
)

var (
	LegacyProposalHandler       = govclient.NewProposalHandler(cli.NewCmdSubmitLegacyUpgradeProposal)
	LegacyCancelProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitLegacyCancelUpgradeProposal)
)
