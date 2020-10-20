package sdk

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk/types"
)

func (c Client) GetNodeInfo(ctx context.Context) (rpc.NodeInfoResponse, error) {
	r, err := http.Get(c.getEndpoint("/node_info"))
	if err != nil {
		return rpc.NodeInfoResponse{}, err
	}
	if r == nil {
		return rpc.NodeInfoResponse{}, fmt.Errorf("unable to fetch data from endpoint %s", c.getEndpoint("/node_info"))
	}
	defer r.Body.Close()
	btes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return rpc.NodeInfoResponse{}, err
	}

	var infoRes rpc.NodeInfoResponse
	if err = types.Codec.UnmarshalJSON(btes, &infoRes); err != nil {
		return infoRes, err
	}

	return infoRes, nil
}
