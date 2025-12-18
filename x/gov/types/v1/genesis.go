package v1

import (
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// NewGenesisState creates a new genesis state for the governance module
func NewGenesisState(startingProposalID uint64, participationEma,
	constitutionParticipationEma, lawParticipationEma string, params Params,
) *GenesisState {
	return &GenesisState{
		StartingProposalId:                    startingProposalID,
		ParticipationEma:                      participationEma,
		ConstitutionAmendmentParticipationEma: constitutionParticipationEma,
		LawParticipationEma:                   lawParticipationEma,
		Params:                                &params,
	}
}

// DefaultGenesisState defines the default governance genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultStartingProposalID,
		DefaultParticipationEma,
		DefaultParticipationEma,
		DefaultParticipationEma,
		DefaultParams(),
	)
}

// Empty returns true if a GenesisState is empty
func (data GenesisState) Empty() bool {
	return data.StartingProposalId == 0 || data.Params == nil
}

// ValidateGenesis checks if gov genesis state is valid ranges
// It checks if params are in valid ranges
// It also makes sure that the provided proposal IDs are unique and
// that there are no duplicate deposit or vote records and no vote or deposits for non-existent proposals
func ValidateGenesis(data *GenesisState) error {
	if data.StartingProposalId == 0 {
		return errors.New("starting proposal id must be greater than 0")
	}

	var errGroup errgroup.Group

	// weed out duplicate proposals
	proposalIDs := make(map[uint64]struct{})
	for _, p := range data.Proposals {
		if _, ok := proposalIDs[p.Id]; ok {
			return fmt.Errorf("duplicate proposal id: %d", p.Id)
		}

		proposalIDs[p.Id] = struct{}{}
	}

	// weed out duplicate deposits
	errGroup.Go(func() error {
		type depositKey struct {
			proposalID uint64
			Depositor  string
		}
		depositIDs := make(map[depositKey]struct{})
		for _, d := range data.Deposits {
			if _, ok := proposalIDs[d.ProposalId]; !ok {
				return fmt.Errorf("deposit %v has non-existent proposal id: %d", d, d.ProposalId)
			}

			dk := depositKey{d.ProposalId, d.Depositor}
			if _, ok := depositIDs[dk]; ok {
				return fmt.Errorf("duplicate deposit: %v", d)
			}

			depositIDs[dk] = struct{}{}
		}

		return nil
	})

	// weed out duplicate votes
	errGroup.Go(func() error {
		type voteKey struct {
			proposalID uint64
			Voter      string
		}
		voteIDs := make(map[voteKey]struct{})
		for _, v := range data.Votes {
			if _, ok := proposalIDs[v.ProposalId]; !ok {
				return fmt.Errorf("vote %v has non-existent proposal id: %d", v, v.ProposalId)
			}

			vk := voteKey{v.ProposalId, v.Voter}
			if _, ok := voteIDs[vk]; ok {
				return fmt.Errorf("duplicate vote: %v", v)
			}

			voteIDs[vk] = struct{}{}
		}

		return nil
	})

	// weed out duplicate governors
	errGroup.Go(func() error {
		governorIDs := make(map[string]struct{})
		for _, g := range data.Governors {
			if _, err := types.GovernorAddressFromBech32(g.GovernorAddress); err != nil {
				return fmt.Errorf("invalid governor address: %v", g)
			}
			if _, ok := governorIDs[g.GovernorAddress]; ok {
				return fmt.Errorf("duplicate governor: %v", g)
			}

			governorIDs[g.GovernorAddress] = struct{}{}
		}

		return nil
	})

	// weed out duplicate governance delegations
	errGroup.Go(func() error {
		delegatorIDs := make(map[string]struct{})
		for _, d := range data.GovernanceDelegations {
			if _, err := sdk.AccAddressFromBech32(d.DelegatorAddress); err != nil {
				return fmt.Errorf("invalid delegator address: %v", d)
			}
			if _, err := types.GovernorAddressFromBech32(d.GovernorAddress); err != nil {
				return fmt.Errorf("invalid governor address: %v", d)
			}
			if _, ok := delegatorIDs[d.DelegatorAddress]; ok {
				return fmt.Errorf("duplicate governance delegation: %v", d)
			}

			delegatorIDs[d.DelegatorAddress] = struct{}{}
		}

		return nil
	})

	// verify params
	errGroup.Go(func() error {
		return data.Params.ValidateBasic()
	})

	return errGroup.Wait()
}

var _ codectypes.UnpackInterfacesMessage = GenesisState{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (data GenesisState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, p := range data.Proposals {
		err := p.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}
