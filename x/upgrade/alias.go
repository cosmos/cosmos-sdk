package upgrade

// nolint

import (
	keeper2 "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	types2 "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	ModuleName                        = types2.ModuleName
	RouterKey                         = types2.RouterKey
	StoreKey                          = types2.StoreKey
	QuerierKey                        = types2.QuerierKey
	PlanByte                          = types2.PlanByte
	DoneByte                          = types2.DoneByte
	ProposalTypeSoftwareUpgrade       = types2.ProposalTypeSoftwareUpgrade
	ProposalTypeCancelSoftwareUpgrade = types2.ProposalTypeCancelSoftwareUpgrade
	QueryCurrent                      = types2.QueryCurrent
	QueryApplied                      = types2.QueryApplied
)

var (
	// functions aliases
	RegisterCodec                    = types2.RegisterCodec
	PlanKey                          = types2.PlanKey
	NewSoftwareUpgradeProposal       = types2.NewSoftwareUpgradeProposal
	NewCancelSoftwareUpgradeProposal = types2.NewCancelSoftwareUpgradeProposal
	NewQueryAppliedParams            = types2.NewQueryAppliedParams
	NewKeeper                        = keeper2.NewKeeper
	NewQuerier                       = keeper2.NewQuerier
)

type (
	UpgradeHandler                = types2.UpgradeHandler // nolint
	Plan                          = types2.Plan
	SoftwareUpgradeProposal       = types2.SoftwareUpgradeProposal
	CancelSoftwareUpgradeProposal = types2.CancelSoftwareUpgradeProposal
	QueryAppliedParams            = types2.QueryAppliedParams
	Keeper                        = keeper2.Keeper
)
