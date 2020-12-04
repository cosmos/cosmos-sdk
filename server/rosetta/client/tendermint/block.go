package tendermint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// tmRawBlock defines the tendermint jsonified raw block
type tmRawBlock struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		BlockMeta struct {
			BlockID struct {
				Hash  string `json:"hash"`
				Parts struct {
					Total string `json:"total"`
					Hash  string `json:"hash"`
				} `json:"parts"`
			} `json:"block_id"`
			Header struct {
				Version struct {
					Block string `json:"block"`
					App   string `json:"app"`
				} `json:"version"`
				ChainID     string    `json:"chain_id"`
				Height      string    `json:"height"`
				Time        time.Time `json:"time"`
				NumTxs      string    `json:"num_txs"`
				TotalTxs    string    `json:"total_txs"`
				LastBlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"last_block_id"`
				LastCommitHash     string `json:"last_commit_hash"`
				DataHash           string `json:"data_hash"`
				ValidatorsHash     string `json:"validators_hash"`
				NextValidatorsHash string `json:"next_validators_hash"`
				ConsensusHash      string `json:"consensus_hash"`
				AppHash            string `json:"app_hash"`
				LastResultsHash    string `json:"last_results_hash"`
				EvidenceHash       string `json:"evidence_hash"`
				ProposerAddress    string `json:"proposer_address"`
			} `json:"header"`
		} `json:"block_meta"`
		Block struct {
			Header struct {
				Version struct {
					Block string `json:"block"`
					App   string `json:"app"`
				} `json:"version"`
				ChainID     string    `json:"chain_id"`
				Height      string    `json:"height"`
				Time        time.Time `json:"time"`
				NumTxs      string    `json:"num_txs"`
				TotalTxs    string    `json:"total_txs"`
				LastBlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"last_block_id"`
				LastCommitHash     string `json:"last_commit_hash"`
				DataHash           string `json:"data_hash"`
				ValidatorsHash     string `json:"validators_hash"`
				NextValidatorsHash string `json:"next_validators_hash"`
				ConsensusHash      string `json:"consensus_hash"`
				AppHash            string `json:"app_hash"`
				LastResultsHash    string `json:"last_results_hash"`
				EvidenceHash       string `json:"evidence_hash"`
				ProposerAddress    string `json:"proposer_address"`
			} `json:"header"`
			Data struct {
				Txs interface{} `json:"txs"`
			} `json:"data"`
			Evidence struct {
				Evidence interface{} `json:"evidence"`
			} `json:"evidence"`
			LastCommit struct {
				BlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"block_id"`
				Precommits []struct {
					Type    int    `json:"type"`
					Height  string `json:"height"`
					Round   string `json:"round"`
					BlockID struct {
						Hash  string `json:"hash"`
						Parts struct {
							Total string `json:"total"`
							Hash  string `json:"hash"`
						} `json:"parts"`
					} `json:"block_id"`
					Timestamp        time.Time `json:"timestamp"`
					ValidatorAddress string    `json:"validator_address"`
					ValidatorIndex   string    `json:"validator_index"`
					Signature        string    `json:"signature"`
				} `json:"precommits"`
			} `json:"last_commit"`
		} `json:"block"`
	} `json:"result"`
}

func (t tmRawBlock) toBlockResponse() BlockResponse {
	return BlockResponse{
		BlockID: BlockID{
			Hash: t.Result.BlockMeta.BlockID.Hash,
		},
		Block: Block{
			Header: BlockHeader{
				LastBlockID: BlockID{Hash: t.Result.Block.Header.LastBlockID.Hash},
				Height:      t.Result.Block.Header.Height,
				Time:        t.Result.Block.Header.Time.String(),
			},
		},
	}
}

type BlockResponse struct {
	BlockID BlockID `json:"block_id,omitempty"`
	Block   Block   `json:"block,omitempty"`
}

type BlockID struct {
	Hash string `json:"hash"`
}

type Block struct {
	Header BlockHeader `json:"header,omitempty"`
}

type BlockHeader struct {
	LastBlockID BlockID `json:"last_block_id"`
	Height      string  `json:"height"`
	Time        string  `json:"time"`
}

func (c Client) Block(height uint64) (BlockResponse, error) {
	var endpoint string
	if height == 0 {
		endpoint = c.getEndpoint("block")
	} else {
		endpoint = c.getEndpoint(fmt.Sprintf("block?height=%d", height))
	}

	resp, err := http.Get(endpoint) // nolint
	if err != nil {
		return BlockResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BlockResponse{}, err
	}
	var rawBlock tmRawBlock
	err = json.Unmarshal(body, &rawBlock)
	if err != nil {
		return BlockResponse{}, err
	}

	return rawBlock.toBlockResponse(), nil
}

func (c Client) BlockByHash(hash string) (BlockResponse, error) {
	resp, err := http.Get(c.getEndpoint(fmt.Sprintf("block_by_hash?hash=%s", hash)))
	if err != nil {
		return BlockResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BlockResponse{}, err
	}

	var rawBlock tmRawBlock
	err = json.Unmarshal(body, &rawBlock)
	if err != nil {
		return BlockResponse{}, err
	}

	return rawBlock.toBlockResponse(), nil
}
