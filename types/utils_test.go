package types_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type utilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(utilsTestSuite))
}

func (s *utilsTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *utilsTestSuite) TestSortJSON() {
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
		{unsortedJSON: `{"consensus_params":{"block_size_params":{"max_bytes":22020096,"max_txs":100000,"max_gas":-1},"tx_size_params":{"max_bytes":10240,"max_gas":-1},"block_gossip_params":{"block_part_size_bytes":65536},"evidence_params":{"max_age":100000}},"validators":[{"pub_key":{"type":"AC26791624DE60","value":"c7UMMAbjFuc5GhGPy0E5q5tefy12p9Tq0imXqdrKXwo="},"power":100,"name":""}],"app_hash":"","genesis_time":"2018-05-11T15:52:25.424795506Z","chain_id":"test-chain-Q6VeoW","app_state":{"accounts":[{"address":"718C9C23F98C9642569742ADDD9F9AB9743FBD5D","coins":[{"denom":"Token","amount":1000},{"denom":"stake","amount":50}]}],"stake":{"pool":{"total_supply":50,"bonded_shares":"0","unbonded_shares":"0","bonded_pool":0,"unbonded_pool":0,"inflation_last_time":0,"inflation":"7/100"},"params":{"inflation_rate_change":"13/100","inflation_max":"1/5","inflation_min":"7/100","goal_bonded":"67/100","max_validators":100,"bond_denom":"stake"},"candidates":null,"bonds":null}}}`,
			want:    `{"app_hash":"","app_state":{"accounts":[{"address":"718C9C23F98C9642569742ADDD9F9AB9743FBD5D","coins":[{"amount":1000,"denom":"Token"},{"amount":50,"denom":"stake"}]}],"stake":{"bonds":null,"candidates":null,"params":{"bond_denom":"stake","goal_bonded":"67/100","inflation_max":"1/5","inflation_min":"7/100","inflation_rate_change":"13/100","max_validators":100},"pool":{"bonded_pool":0,"bonded_shares":"0","inflation":"7/100","inflation_last_time":0,"total_supply":50,"unbonded_pool":0,"unbonded_shares":"0"}}},"chain_id":"test-chain-Q6VeoW","consensus_params":{"block_gossip_params":{"block_part_size_bytes":65536},"block_size_params":{"max_bytes":22020096,"max_gas":-1,"max_txs":100000},"evidence_params":{"max_age":100000},"tx_size_params":{"max_bytes":10240,"max_gas":-1}},"genesis_time":"2018-05-11T15:52:25.424795506Z","validators":[{"name":"","power":100,"pub_key":{"type":"AC26791624DE60","value":"c7UMMAbjFuc5GhGPy0E5q5tefy12p9Tq0imXqdrKXwo="}}]}`,
			wantErr: false},
		// from the TXSpec:
		{unsortedJSON: `{"chain_id":"test-chain-1","sequence":1,"fee_bytes":{"amount":[{"amount":5,"denom":"photon"}],"gas":10000},"msg_bytes":{"inputs":[{"address":"696E707574","coins":[{"amount":10,"denom":"atom"}]}],"outputs":[{"address":"6F7574707574","coins":[{"amount":10,"denom":"atom"}]}]},"alt_bytes":null}`,
			want:    `{"alt_bytes":null,"chain_id":"test-chain-1","fee_bytes":{"amount":[{"amount":5,"denom":"photon"}],"gas":10000},"msg_bytes":{"inputs":[{"address":"696E707574","coins":[{"amount":10,"denom":"atom"}]}],"outputs":[{"address":"6F7574707574","coins":[{"amount":10,"denom":"atom"}]}]},"sequence":1}`,
			wantErr: false},
	}

	for tcIndex, tc := range cases {
		tc := tc
		got, err := sdk.SortJSON([]byte(tc.unsortedJSON))
		if tc.wantErr {
			s.Require().NotNil(err, "tc #%d", tcIndex)
			s.Require().Panics(func() { sdk.MustSortJSON([]byte(tc.unsortedJSON)) })
		} else {
			s.Require().Nil(err, "tc #%d, err=%s", tcIndex, err)
			s.Require().NotPanics(func() { sdk.MustSortJSON([]byte(tc.unsortedJSON)) })
			s.Require().Equal(got, sdk.MustSortJSON([]byte(tc.unsortedJSON)))
		}

		s.Require().Equal(string(got), tc.want)
	}
}

func (s *utilsTestSuite) TestTimeFormatAndParse() {
	cases := []struct {
		RFC3339NanoStr     string
		SDKSortableTimeStr string
		Equal              bool
	}{
		{"2009-11-10T23:00:00Z", "2009-11-10T23:00:00.000000000", true},
		{"2011-01-10T23:10:05.758230235Z", "2011-01-10T23:10:05.758230235", true},
	}
	for _, tc := range cases {
		tc := tc
		timeFromRFC, err := time.Parse(time.RFC3339Nano, tc.RFC3339NanoStr)
		s.Require().Nil(err)
		timeFromSDKFormat, err := time.Parse(sdk.SortableTimeFormat, tc.SDKSortableTimeStr)
		s.Require().Nil(err)

		s.Require().True(timeFromRFC.Equal(timeFromSDKFormat))
		s.Require().Equal(timeFromRFC.Format(sdk.SortableTimeFormat), tc.SDKSortableTimeStr)
	}
}

func (s *utilsTestSuite) TestCopyBytes() {
	s.Require().Nil(sdk.CopyBytes(nil))
	s.Require().Equal(0, len(sdk.CopyBytes([]byte{})))
	bs := []byte("test")
	bsCopy := sdk.CopyBytes(bs)
	s.Require().True(bytes.Equal(bs, bsCopy))
}

func (s *utilsTestSuite) TestUint64ToBigEndian() {
	s.Require().Equal([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, sdk.Uint64ToBigEndian(uint64(0)))
	s.Require().Equal([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa}, sdk.Uint64ToBigEndian(uint64(10)))
}

func (s *utilsTestSuite) TestFormatTimeBytes() {
	tm, err := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Mar 3, 2020 at 7:54pm (UTC)")
	s.Require().NoError(err)
	s.Require().Equal("2020-03-03T19:54:00.000000000", string(sdk.FormatTimeBytes(tm)))
}

func (s *utilsTestSuite) TestParseTimeBytes() {
	tm, err := sdk.ParseTimeBytes([]byte("2020-03-03T19:54:00.000000000"))
	s.Require().NoError(err)
	s.Require().True(tm.Equal(time.Date(2020, 3, 3, 19, 54, 0, 0, time.UTC)))

	_, err = sdk.ParseTimeBytes([]byte{})
	s.Require().Error(err)
}
