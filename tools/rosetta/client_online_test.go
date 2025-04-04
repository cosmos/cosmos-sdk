package rosetta

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegex(t *testing.T) {
	genesisChuck := base64.StdEncoding.EncodeToString([]byte(`"genesis_time":"2021-09-28T09:00:00Z","chain_id":"bombay-12","initial_height":"5900001","consensus_params":{"block":{"max_bytes":"5000000","max_gas":"1000000000","time_iota_ms":"1000"},"evidence":{"max_age_num_blocks":"100000","max_age_duration":"172800000000000","max_bytes":"50000"},"validator":{"pub_key_types":["ed25519"]},"version":{}},"validators":[{"address":"EEA4891F5F8D523A6B4B3EAC84B5C08655A00409","pub_key":{"type":"tendermint/PubKeyEd25519","value":"UX71gTBNumQq42qRd6j/K8XN/y3/HAcuAJxj97utawI="},"power":"60612","name":"BTC.Secure"},{"address":"973F589DE1CC8A54ABE2ABE0E0A4ABF13A9EBAE4","pub_key":{"type":"tendermint/PubKeyEd25519","value":"AmGQvQSAAXzSIscx/6o4rVdRMT9QvairQHaCXsWhY+c="},"power":"835","name":"MoonletWallet"},{"address":"831F402BDA0C9A3F260D4F221780BC22A4C3FB23","pub_key":{"type":"tendermint/PubKeyEd25519","value":"Tw8yKbPNEo113ZNbJJ8joeXokoMdBoazRTwb1NQ77WA="},"power":"102842","name":"BlockNgine"},{"address":"F2683F267D2B4C8714B44D68612DB37A8DD2EED7","pub_key":{"type":"tendermint/PubKeyEd25519","value":"PVE4IcWDE6QEqJSEkx55IDkg5zxBo8tVRzKFMJXYFSQ="},"power":"23200","name":"Luna Station 88"},{"address":"9D2428CBAC68C654BE11BE405344C560E6A0F626","pub_key":{"type":"tendermint/PubKeyEd25519","value":"93hzGmZjPRqOnQkb8BULjqanW3M2p1qIcLVTGkf1Zhk="},"power":"35420","name":"Terra-India"},{"address":"DC9897F22E74BF1B66E2640FA461F785F9BA7627","pub_key":{"type":"tendermint/PubKeyEd25519","value":"mlYb/Dzqwh0YJjfH59OZ4vtp+Zhdq5Oj5MNaGHq1X0E="},"power":"25163","name":"SolidStake"},{"address":"AA1A027E270A2BD7AF154999E6DE9D39C5711DE7","pub_key":{"type":"tendermint/PubKeyEd25519","value":"28z8FlpbC7sR0f1Q8OWFASDNi0FAmdldzetwQ07JJzg="},"power":"34529","name":"syncnode"},{"address":"E548735750DC5015ADDE3B0E7A1294C3B868680B","pub_key":{"type":"tendermint/PubKeyEd25519","value":"BTDtLSKp4wpQrWBwmGvp9isWC5jXaAtX1nrJtsCEWew="},"power":"36082","name":"OneStar"}`))
	height, err := extractInitialHeightFromGenesisChunk(genesisChuck)
	require.NoError(t, err)
	require.Equal(t, height, int64(5900001))
}
