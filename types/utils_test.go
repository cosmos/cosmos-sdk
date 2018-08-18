package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortJSON(t *testing.T) {
	cases := []struct {
		unsortedJSON string
		want         string
		wantErr      bool
	}{
		// simple case
		{unsortedJSON: `{"cosmos":"foo", "atom":"bar",  "tendermint":"foobar"}`,
			want: `{"atom":"bar","cosmos":"foo","tendermint":"foobar"}`, wantErr: false},
		// failing case (invalid JSON):
		{unsortedJSON: `"cosmos":"foo",,,, "atom":"bar",  "tendermint":"foobar"}`,
			want:    "",
			wantErr: true},
		// genesis.json
		{unsortedJSON: `{"consensus_params":{"block_size_params":{"max_bytes":22020096,"max_txs":100000,"max_gas":-1},"tx_size_params":{"max_bytes":10240,"max_gas":-1},"block_gossip_params":{"block_part_size_bytes":65536},"evidence_params":{"max_age":100000}},"validators":[{"pub_key":{"type":"AC26791624DE60","value":"c7UMMAbjFuc5GhGPy0E5q5tefy12p9Tq0imXqdrKXwo="},"power":100,"name":""}],"app_hash":"","genesis_time":"2018-05-11T15:52:25.424795506Z","chain_id":"test-chain-Q6VeoW","app_state":{"accounts":[{"address":"718C9C23F98C9642569742ADDD9F9AB9743FBD5D","coins":[{"denom":"Token","amount":1000},{"denom":"steak","amount":50}]}],"stake":{"pool":{"total_supply":50,"bonded_shares":"0","unbonded_shares":"0","bonded_pool":0,"unbonded_pool":0,"inflation_last_time":0,"inflation":"7/100"},"params":{"inflation_rate_change":"13/100","inflation_max":"1/5","inflation_min":"7/100","goal_bonded":"67/100","max_validators":100,"bond_denom":"steak"},"candidates":null,"bonds":null}}}`,
			want:    `{"app_hash":"","app_state":{"accounts":[{"address":"718C9C23F98C9642569742ADDD9F9AB9743FBD5D","coins":[{"amount":1000,"denom":"Token"},{"amount":50,"denom":"steak"}]}],"stake":{"bonds":null,"candidates":null,"params":{"bond_denom":"steak","goal_bonded":"67/100","inflation_max":"1/5","inflation_min":"7/100","inflation_rate_change":"13/100","max_validators":100},"pool":{"bonded_pool":0,"bonded_shares":"0","inflation":"7/100","inflation_last_time":0,"total_supply":50,"unbonded_pool":0,"unbonded_shares":"0"}}},"chain_id":"test-chain-Q6VeoW","consensus_params":{"block_gossip_params":{"block_part_size_bytes":65536},"block_size_params":{"max_bytes":22020096,"max_gas":-1,"max_txs":100000},"evidence_params":{"max_age":100000},"tx_size_params":{"max_bytes":10240,"max_gas":-1}},"genesis_time":"2018-05-11T15:52:25.424795506Z","validators":[{"name":"","power":100,"pub_key":{"type":"AC26791624DE60","value":"c7UMMAbjFuc5GhGPy0E5q5tefy12p9Tq0imXqdrKXwo="}}]}`,
			wantErr: false},
		// from the TXSpec:
		{unsortedJSON: `{"chain_id":"test-chain-1","sequence":1,"fee_bytes":{"amount":[{"amount":5,"denom":"photon"}],"gas":10000},"msg_bytes":{"inputs":[{"address":"696E707574","coins":[{"amount":10,"denom":"atom"}]}],"outputs":[{"address":"6F7574707574","coins":[{"amount":10,"denom":"atom"}]}]},"alt_bytes":null}`,
			want:    `{"alt_bytes":null,"chain_id":"test-chain-1","fee_bytes":{"amount":[{"amount":5,"denom":"photon"}],"gas":10000},"msg_bytes":{"inputs":[{"address":"696E707574","coins":[{"amount":10,"denom":"atom"}]}],"outputs":[{"address":"6F7574707574","coins":[{"amount":10,"denom":"atom"}]}]},"sequence":1}`,
			wantErr: false},
	}

	for _, tc := range cases {
		got, err := SortJSON([]byte(tc.unsortedJSON))
		if tc.wantErr != (err != nil) {
			t.Fatalf("got %t, want: %t, err=%s", err != nil, tc.wantErr, err)
		}
		require.Equal(t, string(got), tc.want)
	}
}
