package client

import (
	"encoding/base64"
	"strings"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/spf13/pflag"

	errorsmod "cosmossdk.io/errors"

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
	end = min(limit+start, numObjs)

	if start >= numObjs {
		// page is out of bounds
		return -1, -1
	}

	return start, end
}

// ReadPageRequest reads and builds the necessary page request flags for pagination.
// The --page-key flag is expected to hold the base64-encoded next_key emitted in the
// previous page's response, i.e. exactly what the CLI prints, and is decoded here.
func ReadPageRequest(flagSet *pflag.FlagSet) (*query.PageRequest, error) {
	pageKey, _ := flagSet.GetString(flags.FlagPageKey)
	offset, _ := flagSet.GetUint64(flags.FlagOffset)
	limit, _ := flagSet.GetUint64(flags.FlagLimit)
	countTotal, _ := flagSet.GetBool(flags.FlagCountTotal)
	page, _ := flagSet.GetUint64(flags.FlagPage)
	reverse, _ := flagSet.GetBool(flags.FlagReverse)

	if page > 1 && offset > 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "page and offset cannot be used together")
	}

	if page > 1 {
		offset = (page - 1) * limit
	}

	// "null" is how JSON/YAML output prints the empty next_key of the last page; it
	// decodes as valid base64 to garbage bytes, so it must be rejected before decoding.
	if strings.EqualFold(pageKey, "null") {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"invalid --%s %q: this is how an empty next_key is printed on the last page; there are no further pages", flags.FlagPageKey, pageKey)
	}

	key, err := base64.StdEncoding.DecodeString(pageKey)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"invalid --%s %q: %v; expected the base64-encoded next_key from the previous page response", flags.FlagPageKey, pageKey, err)
	}

	return &query.PageRequest{
		Key:        key,
		Offset:     offset,
		Limit:      limit,
		CountTotal: countTotal,
		Reverse:    reverse,
	}, nil
}

// NewClientFromNode sets up Client implementation that communicates with a CometBFT node over
// JSON RPC and WebSockets
func NewClientFromNode(nodeURI string) (*rpchttp.HTTP, error) {
	return rpchttp.New(nodeURI, "/websocket")
}

// Deprecated: ReadPageRequest now base64-decodes the page-key itself, so this is a
// no-op returning flagSet unchanged. Decoding here as well would leave ReadPageRequest
// with already-raw bytes to decode a second time. Call ReadPageRequest directly.
func FlagSetWithPageKeyDecoded(flagSet *pflag.FlagSet) (*pflag.FlagSet, error) {
	return flagSet, nil
}

// Deprecated: see FlagSetWithPageKeyDecoded. This is a no-op returning flagSet unchanged.
func MustFlagSetWithPageKeyDecoded(flagSet *pflag.FlagSet) *pflag.FlagSet {
	return flagSet
}
