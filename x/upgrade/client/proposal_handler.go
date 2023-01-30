package client

import (
	"cosmossdk.io/x/upgrade/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	LegacyProposalHandler       = govclient.NewProposalHandler(cli.NewCmdSubmitLegacyUpgradeProposal)
	LegacyCancelProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitLegacyCancelUpgradeProposal)
)
