package query

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
)

// Example Code
/*
func (...) GetProposalsFiltered(ctx sdk.Context, params types.QueryProposalsParams) types.Proposals {
	var filteredProposals []types.Proposal

	pageReq := &query.PageRequest{Key: nil, Limit: 10, Offset: 10, CountTotal: true}
	res, err := query.FilteredPaginate(accountStore, request.Req, func(key []byte, value []byte, accumulate bool) (hit, error) {
		var p types.Proposal
		err := app.Codec().UnmarshalBinaryBare(value, &p)
		if err != nil {
			return false, err
		}

		matchVoter, matchDepositor, matchStatus := true, true, true

		// match status (if supplied/valid)
		if types.ValidProposalStatus(params.ProposalStatus) {
			matchStatus = p.Status == params.ProposalStatus
		}

		// match voter address (if supplied)
		if len(params.Voter) > 0 {
			_, matchVoter = keeper.GetVote(ctx, p.ProposalID, params.Voter)
		}

		// match depositor (if supplied)
		if len(params.Depositor) > 0 {
			_, matchDepositor = keeper.GetDeposit(ctx, p.ProposalID, params.Depositor)
		}

		if matchVoter && matchDepositor && matchStatus {
			if accumulate {
				filteredProposals = append(filteredProposals, p)
			}

			return true, nil
		}

		return false, nil
	})

	// handle error

	return types.QueryProposalsResponse{Proposals: filteredProposals, Res: res}
 */

// FilteredPaginate does pagination of all the results in the PrefixStore based on the
// provided PageRequest. onResult should be used to do actual unmarshaling and filter the results.
// if key is provided, the pagination uses the optimized querying
// if offset is used, the pagination uses lazy filtering i.e., searches through all the records
// accumulate represents if the response is valid based on the offset given
// it will be false for the results (filtered) < offset  and true for `offset > accumulate <= end`. This
// will help to append the result to original client response
func FilteredPaginate(
	prefixStore types.KVStore,
	req *PageRequest,
	onResult func(key []byte, value []byte, accumulate bool) (bool, error),
) (*PageResponse, error) {
	offset := req.Offset
	key := req.Key
	limit := req.Limit
	countTotal := req.CountTotal

	if offset > 0 && key != nil {
		return nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	if limit == 0 {
		limit = defaultLimit

		// count total results when the limit is zero/not supplied
		countTotal = true
	}

	if len(key) != 0 {
		iterator := prefixStore.Iterator(key, nil)
		defer iterator.Close()

		var count uint64
		var nextKey []byte

		for ; iterator.Valid(); iterator.Next() {
			if count == limit {
				nextKey = iterator.Key()
				break
			}

			hit, err := onResult(iterator.Key(), iterator.Value(), true)
			if err != nil {
				return nil, err
			}

			if hit {
				count++
			}
		}

		return &PageResponse{
			NextKey: nextKey,
		}, nil
	}

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	end := offset + limit

	var numHits uint64
	var nextKey []byte

	for ; iterator.Valid(); iterator.Next() {
		accumulate := numHits > offset && numHits <= end
		hit, err := onResult(iterator.Key(), iterator.Value(), accumulate)
		if err != nil {
			return nil, err
		}

		if hit {
			numHits++
		}

		if numHits == end {
			nextKey = iterator.Key()

			if !countTotal {
				break
			}
		}
	}

	res := &PageResponse{NextKey: nextKey}
	if countTotal {
		res.Total = numHits
	}

	return res, nil
}
