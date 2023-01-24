package client

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/pflag"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// Paginate returns the correct starting and ending index for a paginated query,
// given that client provides a desired page and limit of objects and the handler
// provides the total number of objects. The start page is assumed to be 1-indexed.
// If the start page is invalid, non-positive values are returned signaling the
// request is invalid; it returns non-positive values if limit is non-positive and
// defLimit is negative.
func Paginate(numObjs, page, limit, defLimit int) (start, end int) {
	if page <= 0 {
		// invalid start page
		return -1, -1
	}

	// fallback to default limit if supplied limit is invalid
	if limit <= 0 {
		if defLimit < 0 {
			// invalid default limit
			return -1, -1
		}
		limit = defLimit
	}

	start = (page - 1) * limit
	end = limit + start

	if end >= numObjs {
		end = numObjs
	}

	if start >= numObjs {
		// page is out of bounds
		return -1, -1
	}

	return start, end
}

// A FlagSetMutator is a function that takes in a flagSet and possibly modifies entries.
type FlagSetMutator = func(flagSet *pflag.FlagSet) (*pflag.FlagSet, error)

// ReadPageRequest reads and builds the necessary page request flags for pagination.
// If one or more mutators are provided, they are applied to the provided flagSet before attempting to read the flags.
func ReadPageRequest(flagSet *pflag.FlagSet, mutators ...FlagSetMutator) (*query.PageRequest, error) {
	var err error
	for _, mutator := range mutators {
		flagSet, err = mutator(flagSet)
		if err != nil {
			return nil, err
		}
	}

	pageKey, _ := flagSet.GetString(flags.FlagPageKey)
	offset, _ := flagSet.GetUint64(flags.FlagOffset)
	limit, _ := flagSet.GetUint64(flags.FlagLimit)
	countTotal, _ := flagSet.GetBool(flags.FlagCountTotal)
	page, _ := flagSet.GetUint64(flags.FlagPage)
	reverse, _ := flagSet.GetBool(flags.FlagReverse)

	if page > 1 && offset > 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "page and offset cannot be used together")
	}

	if page > 1 {
		offset = (page - 1) * limit
	}

	return &query.PageRequest{
		Key:        []byte(pageKey),
		Offset:     offset,
		Limit:      limit,
		CountTotal: countTotal,
		Reverse:    reverse,
	}, nil
}

// ReadPageRequestWithPageKeyDecoded is a shortcut for ReadPageRequest(flagSet, FlagSetWithPageKeyDecoded)
func ReadPageRequestWithPageKeyDecoded(flagSet *pflag.FlagSet) (*query.PageRequest, error) {
	return ReadPageRequest(flagSet, FlagSetWithPageKeyDecoded)
}

// NewClientFromNode sets up Client implementation that communicates with a Tendermint node over
// JSON RPC and WebSockets
func NewClientFromNode(nodeURI string) (*rpchttp.HTTP, error) {
	return rpchttp.New(nodeURI, "/websocket")
}

// FlagSetWithPageKeyDecoded returns the provided flagSet with the page-key value base64 decoded (if it exists).
// This is for when the page-key is provided as a base64 string (e.g. from the CLI).
// ReadPageRequest expects it to be the raw bytes.
//
// Common usage:
// fs, err := client.FlagSetWithPageKeyDecoded(cmd.Flags())
// pageReq, err := client.ReadPageRequest(fs)
func FlagSetWithPageKeyDecoded(flagSet *pflag.FlagSet) (*pflag.FlagSet, error) {
	encoded, err := flagSet.GetString(flags.FlagPageKey)
	if err != nil {
		return flagSet, err
	}
	if len(encoded) > 0 {
		var raw []byte
		raw, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return flagSet, fmt.Errorf("error decoding %s flag: %w", flags.FlagPageKey, err)
		}
		_ = flagSet.Set(flags.FlagPageKey, string(raw))
	}
	return flagSet, nil
}

// MustFlagSetWithPageKeyDecoded calls FlagSetWithPageKeyDecoded and panics on error.
//
// Common usage: pageReq, err := client.ReadPageRequest(client.MustFlagSetWithPageKeyDecoded(cmd.Flags()))
func MustFlagSetWithPageKeyDecoded(flagSet *pflag.FlagSet) *pflag.FlagSet {
	rv, err := FlagSetWithPageKeyDecoded(flagSet)
	if err != nil {
		panic(err.Error())
	}
	return rv
}
