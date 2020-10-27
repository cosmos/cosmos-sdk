package tendermint

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type NetInfoResponse struct {
	NPeers string `json:"n_peers,omitempty"`
	Peers  []Peer `json:"peers,omitempty"`
}

type Peer struct {
	NodeInfo NodeInfo `json:"node_info,omitempty"`
}

type NodeInfo struct {
	ID string `json:"id,omitempty"`
}

func (c Client) NetInfo() (NetInfoResponse, error) {
	resp, err := http.Get(c.getEndpoint("net_info"))
	if err != nil {
		return NetInfoResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NetInfoResponse{}, err
	}

	var jsonResp map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return NetInfoResponse{}, err
	}

	var netInfoResp NetInfoResponse
	err = json.Unmarshal(jsonResp["result"], &netInfoResp)
	if err != nil {
		return NetInfoResponse{}, err
	}

	return netInfoResp, nil
}
