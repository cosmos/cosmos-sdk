package keeper

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
)

func (k Keeper) InitGenesis(ctx types.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState group.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	if err := k.groupTable.Import(ctx.KVStore(k.key), genesisState.Groups, genesisState.GroupSeq); err != nil {
		panic(errors.Wrap(err, "groups"))
	}

	if err := k.groupMemberTable.Import(ctx.KVStore(k.key), genesisState.GroupMembers, 0); err != nil {
		panic(errors.Wrap(err, "group members"))
	}

	if err := k.groupAccountTable.Import(ctx.KVStore(k.key), genesisState.GroupAccounts, 0); err != nil {
		panic(errors.Wrap(err, "group accounts"))
	}

	if err := k.groupAccountSeq.InitVal(ctx.KVStore(k.key), genesisState.GroupAccountSeq); err != nil {
		panic(errors.Wrap(err, "group account seq"))
	}

	if err := k.proposalTable.Import(ctx.KVStore(k.key), genesisState.Proposals, genesisState.ProposalSeq); err != nil {
		panic(errors.Wrap(err, "proposals"))
	}

	if err := k.voteTable.Import(ctx.KVStore(k.key), genesisState.Votes, 0); err != nil {
		panic(errors.Wrap(err, "votes"))
	}

	return []abci.ValidatorUpdate{}

}

func (k Keeper) ExportGenesis(ctx types.Context, cdc codec.JSONCodec) json.RawMessage {
	genesisState := group.NewGenesisState()

	var groups []*group.GroupInfo

	groupSeq, err := k.groupTable.Export(ctx.KVStore(k.key), &groups)
	if err != nil {
		panic(errors.Wrap(err, "groups"))
	}
	genesisState.Groups = groups
	genesisState.GroupSeq = groupSeq

	var groupMembers []*group.GroupMember
	_, err = k.groupMemberTable.Export(ctx.KVStore(k.key), &groupMembers)
	if err != nil {
		panic(errors.Wrap(err, "group members"))
	}
	genesisState.GroupMembers = groupMembers

	var groupAccounts []*group.GroupAccountInfo
	_, err = k.groupAccountTable.Export(ctx.KVStore(k.key), &groupAccounts)
	if err != nil {
		panic(errors.Wrap(err, "group accounts"))
	}
	genesisState.GroupAccounts = groupAccounts
	genesisState.GroupAccountSeq = k.groupAccountSeq.CurVal(ctx.KVStore(k.key))

	var proposals []*group.Proposal
	proposalSeq, err := k.proposalTable.Export(ctx.KVStore(k.key), &proposals)
	if err != nil {
		panic(errors.Wrap(err, "proposals"))
	}
	genesisState.Proposals = proposals
	genesisState.ProposalSeq = proposalSeq

	var votes []*group.Vote
	_, err = k.voteTable.Export(ctx.KVStore(k.key), &votes)
	if err != nil {
		panic(errors.Wrap(err, "votes"))
	}
	genesisState.Votes = votes

	genesisBytes := cdc.MustMarshalJSON(genesisState)
	return genesisBytes
}
