package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/p2p"

	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/version"
)

// aliasNodeInfo is required due to the fact that as of tendermint version 0.33.9
// there seems to be a problem while unmarshalling types using the amino codec
// buildDeps elements is being treated as a struct instead of string by amino
// and hence the unmarshalling fails due to amino trying to unmarshal a string into
// type map[string]json.RawMessage
type aliasedNodeInfo struct {
	NodeInfo struct {
		ProtocolVersion struct {
			P2P   string `json:"p2p"`
			Block string `json:"block"`
			App   string `json:"app"`
		} `json:"protocol_version"`
		ID         string `json:"id"`
		ListenAddr string `json:"listen_addr"`
		Network    string `json:"network"`
		Version    string `json:"version"`
		Channels   string `json:"channels"`
		Moniker    string `json:"moniker"`
		Other      struct {
			TxIndex    string `json:"tx_index"`
			RPCAddress string `json:"rpc_address"`
		} `json:"other"`
	} `json:"node_info"`
	ApplicationVersion struct {
		Name       string   `json:"name"`
		ServerName string   `json:"server_name"`
		ClientName string   `json:"client_name"`
		Version    string   `json:"version"`
		Commit     string   `json:"commit"`
		BuildTags  string   `json:"build_tags"`
		Go         string   `json:"go"`
		BuildDeps  []string `json:"build_deps"`
	} `json:"application_version"`
}

func (c Client) GetNodeInfo(ctx context.Context) (resp rpc.NodeInfoResponse, err error) {
	defer func() {
		if err != nil {
			log.Printf("%s", err)
		}
	}()
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

	var infoRes aliasedNodeInfo
	if err = json.Unmarshal(btes, &infoRes); err != nil {
		return rpc.NodeInfoResponse{}, err
	}
	return rpc.NodeInfoResponse{
		DefaultNodeInfo: p2p.DefaultNodeInfo{
			ProtocolVersion: p2p.ProtocolVersion{},
			DefaultNodeID:   (p2p.ID)(infoRes.NodeInfo.ID),
			ListenAddr:      infoRes.NodeInfo.ListenAddr,
			Network:         infoRes.NodeInfo.Network,
			Version:         infoRes.NodeInfo.Version,
			Channels:        bytes.HexBytes(infoRes.NodeInfo.Channels),
			Moniker:         infoRes.NodeInfo.Moniker,
			Other: p2p.DefaultNodeInfoOther{
				TxIndex:    infoRes.NodeInfo.Other.TxIndex,
				RPCAddress: infoRes.NodeInfo.Other.RPCAddress,
			},
		},
		ApplicationVersion: version.Info{
			Name:       infoRes.ApplicationVersion.Name,
			ServerName: infoRes.ApplicationVersion.ServerName,
			ClientName: infoRes.ApplicationVersion.ClientName,
			Version:    infoRes.ApplicationVersion.Version,
			GitCommit:  infoRes.ApplicationVersion.Commit,
			BuildTags:  infoRes.ApplicationVersion.BuildTags,
			GoVersion:  infoRes.ApplicationVersion.Go,
			// BuildDeps:  infoRes.ApplicationVersion.BuildDeps,
		},
	}, nil
}
