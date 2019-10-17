package rpc

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/assert"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var defaultBlockJson = `{"block_meta":{"block_id":{"hash":"30CD3A9EF2082FF9F2575655E99E37ACC3936DC9E534ADF0BD7436C76258225C","parts":{"total":"1","hash":"463460DA73FA0AE441B6041BF2683AE625CE2D9FD290F4C86CD83D1A7FB9F439"}},"header":{"version":{"block":"10","app":"0"},"chain_id":"test-chain-RoCfX4","height":"16","time":"2019-10-16T09:47:21.054275229Z","num_txs":"0","total_txs":"0","last_block_id":{"hash":"168D140232DA3A59757658E257168B6BCE62024F574CA222C2D449BAB1AE4023","parts":{"total":"1","hash":"4E6D7A1C7DE6942470394C21B132A2D9F89B31908C3B6FA8C2AB96A87553B96E"}},"last_commit_hash":"3483509FCC8048496E9B85E1E4C90A3B6E5944F022FE3C0709096932A010F712","data_hash":"","validators_hash":"61E69204249E5EE6F777C0566040929276A1900CE665BCB11006CABF9A4ED9A3","next_validators_hash":"61E69204249E5EE6F777C0566040929276A1900CE665BCB11006CABF9A4ED9A3","consensus_hash":"048091BC7DDC283F77BFBF91D73C44DA58C3DF8A9CBC867405D8B7F3DAADA22F","app_hash":"070A8213F67F494F7D27A6296C29F0DF62DFA18FD37FB592D9761A44DEE96528","last_results_hash":"","evidence_hash":"","proposer_address":"E8B4AF895B301C3D7108CA93A0B9234E3A2A4B21"}},"block":{"header":{"version":{"block":"10","app":"0"},"chain_id":"test-chain-RoCfX4","height":"16","time":"2019-10-16T09:47:21.054275229Z","num_txs":"0","total_txs":"0","last_block_id":{"hash":"168D140232DA3A59757658E257168B6BCE62024F574CA222C2D449BAB1AE4023","parts":{"total":"1","hash":"4E6D7A1C7DE6942470394C21B132A2D9F89B31908C3B6FA8C2AB96A87553B96E"}},"last_commit_hash":"3483509FCC8048496E9B85E1E4C90A3B6E5944F022FE3C0709096932A010F712","data_hash":"","validators_hash":"61E69204249E5EE6F777C0566040929276A1900CE665BCB11006CABF9A4ED9A3","next_validators_hash":"61E69204249E5EE6F777C0566040929276A1900CE665BCB11006CABF9A4ED9A3","consensus_hash":"048091BC7DDC283F77BFBF91D73C44DA58C3DF8A9CBC867405D8B7F3DAADA22F","app_hash":"070A8213F67F494F7D27A6296C29F0DF62DFA18FD37FB592D9761A44DEE96528","last_results_hash":"","evidence_hash":"","proposer_address":"E8B4AF895B301C3D7108CA93A0B9234E3A2A4B21"},"data":{"txs":null},"evidence":{"evidence":null},"last_commit":{"block_id":{"hash":"168D140232DA3A59757658E257168B6BCE62024F574CA222C2D449BAB1AE4023","parts":{"total":"1","hash":"4E6D7A1C7DE6942470394C21B132A2D9F89B31908C3B6FA8C2AB96A87553B96E"}},"precommits":[{"type":2,"height":"15","round":"0","block_id":{"hash":"168D140232DA3A59757658E257168B6BCE62024F574CA222C2D449BAB1AE4023","parts":{"total":"1","hash":"4E6D7A1C7DE6942470394C21B132A2D9F89B31908C3B6FA8C2AB96A87553B96E"}},"timestamp":"2019-10-16T09:47:21.054275229Z","validator_address":"E8B4AF895B301C3D7108CA93A0B9234E3A2A4B21","validator_index":"0","signature":"eXB1NPmarUC5+2SCUQiWUCoGqtq0nrSn7giMAF/zWbfv3KmueE+1NoZ0CPLJAhBodz8Y8oL8xVy6ElA7J72kBw=="}]}}}`

func TestConvertBlockResult(t *testing.T) {
	cdc := codec.New()

	var block ctypes.ResultBlock
	cdc.MustUnmarshalJSON([]byte(defaultBlockJson), &block)

	convertedBlock := ConvertBlockResult(&block)
	assert.Equal(t, "cosmosvaloper1az62lz2mxqwr6ugge2f6pwfrfcaz5jephqhvkt", convertedBlock.BlockMeta.Header.ProposerAddress.String())

	convertedBlockString := string(cdc.MustMarshalJSON(&convertedBlock))
	assert.NotContains(t, "E8B4AF895B301C3D7108CA93A0B9234E3A2A4B21", convertedBlockString)
}
