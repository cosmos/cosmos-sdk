package v046

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
)

func convertProposal(oldProp v1beta1.Proposal) (v1beta2.Proposal, error) {
	msg, err := v1beta2.NewLegacyContent(oldProp.GetContent(), authtypes.NewModuleAddress(ModuleName).String())
	if err != nil {
		return v1beta2.Proposal{}, err
	}
	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return v1beta2.Proposal{}, err
	}

	return v1beta2.Proposal{
		ProposalId: oldProp.ProposalId,
		Messages:   []*codectypes.Any{msgAny},
		Status:     v1beta2.ProposalStatus(oldProp.Status),
		FinalTallyResult: &v1beta2.TallyResult{
			Yes:        oldProp.FinalTallyResult.Yes.String(),
			No:         oldProp.FinalTallyResult.No.String(),
			Abstain:    oldProp.FinalTallyResult.Abstain.String(),
			NoWithVeto: oldProp.FinalTallyResult.NoWithVeto.String(),
		},
		SubmitTime:      &oldProp.SubmitTime,
		DepositEndTime:  &oldProp.DepositEndTime,
		TotalDeposit:    oldProp.TotalDeposit,
		VotingStartTime: &oldProp.VotingStartTime,
		VotingEndTime:   &oldProp.VotingEndTime,
	}, nil
}

// migrateProposals migrates all legacy proposals into MsgExecLegacyContent
// proposals.
func migrateProposals(store sdk.KVStore, cdc codec.BinaryCodec) error {
	propStore := prefix.NewStore(store, ProposalsKeyPrefix)

	iter := propStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var oldProp v1beta1.Proposal
		err := cdc.Unmarshal(iter.Value(), &oldProp)
		if err != nil {
			return err
		}

		newProp, err := convertProposal(oldProp)
		if err != nil {
			return err
		}
		bz, err := cdc.Marshal(&newProp)
		if err != nil {
			return err
		}

		// Set new value on store.
		store.Set(iter.Key(), bz)
	}

	return nil
}

// MigrateStore performs in-place store migrations from v0.40 to v0.43. The
// migration includes:
//
// - Migrate proposals to be Msg-based.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	return migrateProposals(store, cdc)
}
