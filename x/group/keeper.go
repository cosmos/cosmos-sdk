package group

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/delegate"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bech32"
	"golang.org/x/crypto/blake2b"
)

type Keeper struct {
	groupStoreKey sdk.StoreKey
	cdc           *codec.Codec
	accountKeeper auth.AccountKeeper
	dispatcher  delegate.Dispatcher
}

func NewKeeper(groupStoreKey sdk.StoreKey, cdc *codec.Codec, accountKeeper auth.AccountKeeper, dispatcher delegate.Dispatcher) Keeper {
	return Keeper{
		groupStoreKey,
		cdc,
		accountKeeper,
		dispatcher,
	}
}

type GroupAccount struct {
	*auth.BaseAccount
}

func (acc *GroupAccount) SetPubKey(pubKey crypto.PubKey) error {
	return fmt.Errorf("cannot set a PubKey on a Group account")
}

var (
	keyNewGroupID = []byte("newGroupID")
)

func KeyGroupID(id sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("g/%d", id))
}

func KeyProposal(id []byte) []byte {
	return []byte(fmt.Sprintf("p/%d", id))
}

func (keeper Keeper) GetGroupInfo(ctx sdk.Context, id sdk.AccAddress) (info Group, err sdk.Error) {
	if len(id) < 1 || id[0] != 'G' {
		return info, sdk.ErrUnknownRequest("Not a valid group")
	}
	store := ctx.KVStore(keeper.groupStoreKey)
	bz := store.Get(KeyGroupID(id))
	if bz == nil {
		return info, sdk.ErrUnknownRequest("Not found")
	}
	info = Group{}
	marshalErr := keeper.cdc.UnmarshalBinaryBare(bz, &info)
	if marshalErr != nil {
		return info, sdk.ErrUnknownRequest(marshalErr.Error())
	}
	return info, nil
}

func AddrFromUint64(id uint64) sdk.AccAddress {
	addr := make([]byte, binary.MaxVarintLen64+1)
	addr[0] = 'G'
	n := binary.PutUvarint(addr[1:], id)
	return addr[:n+1]
}

func (keeper Keeper) getNewGroupId(ctx sdk.Context) sdk.AccAddress {
	store := ctx.KVStore(keeper.groupStoreKey)
	bz := store.Get(keyNewGroupID)
	var groupId uint64 = 0
	if bz != nil {
		keeper.cdc.MustUnmarshalBinaryBare(bz, &groupId)
	}
	bz = keeper.cdc.MustMarshalBinaryBare(groupId + 1)
	store.Set(keyNewGroupID, bz)
	return AddrFromUint64(groupId)
}

func (keeper Keeper) CreateGroup(ctx sdk.Context, info Group) (sdk.AccAddress, error) {
	id := keeper.getNewGroupId(ctx)
	keeper.setGroupInfo(ctx, id, info)
	acct := &GroupAccount{
		BaseAccount: &auth.BaseAccount{
			Address: id,
		},
	}
	existingAcc := keeper.accountKeeper.GetAccount(ctx, id)
	if existingAcc != nil {
		return nil, fmt.Errorf("account with address %s already exists", id.String())
	}
	keeper.accountKeeper.SetAccount(ctx, acct)
	return id, nil
}

func (keeper Keeper) setGroupInfo(ctx sdk.Context, id sdk.AccAddress, info Group) {
	store := ctx.KVStore(keeper.groupStoreKey)
	bz, err := keeper.cdc.MarshalBinaryBare(info)
	if err != nil {
		panic(err)
	}
	store.Set(KeyGroupID(id), bz)
}

//func (keeper Keeper) UpdateGroupInfo(ctx sdk.Context, id GroupID, signers []sdk.AccAddress, info GroupInfo) bool {
//	if !keeper.Authorize(ctx, id, signers) {
//		return false
//	}
//	keeper.setGroupInfo(ctx, id, info)
//	return true
//}

func (keeper Keeper) Authorize(ctx sdk.Context, group sdk.AccAddress, signers []sdk.AccAddress) bool {
	info, err := keeper.GetGroupInfo(ctx, group)
	if err != nil {
		return false
	}
	ctx.GasMeter().ConsumeGas(10, "group auth")
	return keeper.AuthorizeGroupInfo(ctx, &info, signers)
}

func (keeper Keeper) AuthorizeGroupInfo(ctx sdk.Context, info *Group, signers []sdk.AccAddress) bool {
	voteCount := sdk.NewInt(0)
	sigThreshold := info.DecisionThreshold

	nMembers := len(info.Members)
	nSigners := len(signers)
	for i := 0; i < nMembers; i++ {
		mem := info.Members[i]
		// TODO Use a hash map to optimize this
		for j := 0; j < nSigners; j++ {
			ctx.GasMeter().ConsumeGas(10, "check addr")
			if bytes.Compare(mem.Address, signers[j]) == 0 || keeper.Authorize(ctx, mem.Address, signers) {
				voteCount = voteCount.Add(mem.Weight)
				diff := voteCount.Sub(sigThreshold)
				if diff.IsZero() || diff.IsPositive() {
					return true
				}
				break
			}
		}
	}
	return false
}

const (
	Bech32Prefix = "proposal"
)

func mustEncodeProposalIDBech32(id []byte) string {
	str, err := bech32.ConvertAndEncode(Bech32Prefix, id)
	if err != nil {
		panic(err)
	}
	return str
}

func MustDecodeProposalIDBech32(bech string) []byte {
	hrp, data, err := bech32.DecodeAndConvert(bech)
	if err != nil {
		panic(err)
	}
	if hrp != Bech32Prefix {
		panic(fmt.Sprintf("Expected bech32 prefix %s", Bech32Prefix))
	}
	return data
}

func (keeper Keeper) Propose(ctx sdk.Context, proposer sdk.AccAddress, action delegate.Action) ([]byte, sdk.Result) {
	// TODO
	//canHandle, res := keeper.handler.CheckProposal(ctx, action)
	//if !canHandle {
	//	return nil, sdk.ErrUnknownRequest("unknown proposal type").Result()
	//}
	//if res.Code != sdk.CodeOK {
	//	return nil, res
	//}

	store := ctx.KVStore(keeper.groupStoreKey)
	hashBz := blake2b.Sum256(action.GetSignBytes())
	id := hashBz[:]
	bech := mustEncodeProposalIDBech32(id)
	if store.Has(id) {
		return id, sdk.ErrUnknownRequest(fmt.Sprintf("proposal %s already exists", bech)).Result()
	}

	prop := Proposal{
		Proposer:  proposer,
		Action:    action,
		Approvers: []sdk.AccAddress{proposer},
	}

	keeper.storeProposal(ctx, id, &prop)

	res := sdk.Result{}
	res.Tags = res.Tags.
		AppendTag("proposal.id", mustEncodeProposalIDBech32(id)).
		AppendTag("proposal.action", action.Type())
	return id, res
}

func (keeper Keeper) storeProposal(ctx sdk.Context, id []byte, proposal *Proposal) {
	store := ctx.KVStore(keeper.groupStoreKey)
	bz, err := keeper.cdc.MarshalBinaryBare(proposal)
	if err != nil {
		panic(err)
	}

	store.Set(KeyProposal(id), bz)
}

func (keeper Keeper) GetProposal(ctx sdk.Context, id []byte) (proposal *Proposal, err sdk.Error) {
	store := ctx.KVStore(keeper.groupStoreKey)
	bz := store.Get(KeyProposal(id))
	proposal = &Proposal{}
	marshalErr := keeper.cdc.UnmarshalBinaryBare(bz, proposal)
	if marshalErr != nil {
		return proposal, sdk.ErrUnknownRequest(marshalErr.Error())
	}
	return proposal, nil
}

func (keeper Keeper) Vote(ctx sdk.Context, proposalId []byte, voter sdk.AccAddress, yesNo bool) sdk.Result {
	proposal, err := keeper.GetProposal(ctx, proposalId)

	if err != nil {
		return sdk.Result{
			Code: sdk.CodeUnknownRequest,
			Log:  "can't find proposal",
		}
	}

	var newVotes []sdk.AccAddress
	votes := proposal.Approvers
	nVotes := len(votes)

	if yesNo {
		newVotes = make([]sdk.AccAddress, nVotes+1)
		for i := 0; i < nVotes; i++ {
			oldVoter := votes[i]
			if bytes.Equal(voter, oldVoter) {
				// Already voted YES
				return sdk.Result{
					Code: sdk.CodeUnknownRequest,
					Log:  "already voted yes",
				}
			}
			newVotes[i] = oldVoter
		}
		newVotes[nVotes] = voter
	} else {
		newVotes = make([]sdk.AccAddress, nVotes)
		didntVote := true
		j := 0
		for i := 0; i < nVotes; i++ {
			oldVoter := votes[i]
			if bytes.Equal(voter, oldVoter) {
				didntVote = false
			} else {
				newVotes[j] = oldVoter
				j++
			}
		}
		if didntVote {
			return sdk.Result{
				Code: sdk.CodeUnknownRequest,
				Log:  "didn't vote yes previously",
			}
		}
		if j != nVotes-1 {
			panic("unexpected vote count")
		}
		newVotes = newVotes[:j]
	}

	newProp := Proposal{
		Proposer:  proposal.Proposer,
		Action:    proposal.Action,
		Approvers: newVotes,
	}

	keeper.storeProposal(ctx, proposalId, &newProp)

	return sdk.Result{Code: sdk.CodeOK,
		Tags: sdk.EmptyTags().
			AppendTag("proposal.id", mustEncodeProposalIDBech32(proposalId)).
			AppendTag("proposal.action", proposal.Action.Type()),
	}
}

func (keeper Keeper) TryExecute(ctx sdk.Context, proposalId []byte) sdk.Result {
	proposal, err := keeper.GetProposal(ctx, proposalId)

	if err != nil {
		return sdk.ErrUnknownRequest("can't find proposal").Result()
	}

	if !keeper.Authorize(ctx, proposal.Group, proposal.Approvers) {
		return sdk.ErrUnauthorized("proposal failed").Result()
	}

	res := keeper.dispatcher.DispatchAction(ctx, proposal.Group, proposal.Action)

	if res.Code == sdk.CodeOK {
		store := ctx.KVStore(keeper.groupStoreKey)
		store.Delete(KeyProposal(proposalId))
		res.Tags = res.Tags.AppendTag("action", proposal.Action.Type())
	}

	return res
}

func (keeper Keeper) Withdraw(ctx sdk.Context, proposalId []byte, proposer sdk.AccAddress) sdk.Result {
	proposal, err := keeper.GetProposal(ctx, proposalId)

	if err != nil {
		return sdk.Result{
			Code: sdk.CodeUnknownRequest,
			Log:  "can't find proposal",
		}
	}

	if !bytes.Equal(proposer, proposal.Proposer) {
		return sdk.Result{
			Code: sdk.CodeUnauthorized,
			Log:  "you didn't propose this",
		}
	}

	store := ctx.KVStore(keeper.groupStoreKey)
	store.Delete(KeyProposal(proposalId))

	return sdk.Result{Code: sdk.CodeOK,
		Tags: sdk.EmptyTags().
			AppendTag("proposal.id", mustEncodeProposalIDBech32(proposalId)),
	}
}
