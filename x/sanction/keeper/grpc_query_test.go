package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

type QueryTestSuite struct {
	BaseTestSuite
}

func (s *QueryTestSuite) SetupTest() {
	s.BaseSetup()
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

func (s *QueryTestSuite) TestKeeper_IsSanctioned() {
	addrNotSanctioned := sdk.AccAddress("not_sanctioned_addr")
	addrSanctioned := sdk.AccAddress("sanctioned_address")
	AddrTempSanct := sdk.AccAddress("temporarily_sanctioned_address")
	addrSanctTempUn := sdk.AccAddress("this_address_is_nuts")

	s.ClearState()
	s.ReqOKAddPermSanct("addrSanctioned, addrSanctTempUn", addrSanctioned, addrSanctTempUn)
	s.ReqOKAddTempSanct(1, "AddrTempSanct", AddrTempSanct)
	s.ReqOKAddTempUnsanct(1, "addrSanctTempUn", addrSanctTempUn)

	tests := []struct {
		name   string
		req    *sanction.QueryIsSanctionedRequest
		exp    *sanction.QueryIsSanctionedResponse
		expErr []string
	}{
		{
			name:   "nil req",
			req:    nil,
			expErr: []string{"InvalidArgument", "empty request"},
		},
		{
			name:   "no address",
			req:    &sanction.QueryIsSanctionedRequest{Address: ""},
			expErr: []string{"InvalidArgument", "address cannot be empty"},
		},
		{
			name:   "bad address",
			req:    &sanction.QueryIsSanctionedRequest{Address: "not1addr"},
			expErr: []string{"invalid address", "InvalidArgument", "decoding bech32 failed"},
		},
		{
			name: "normal address",
			req:  &sanction.QueryIsSanctionedRequest{Address: addrNotSanctioned.String()},
			exp:  &sanction.QueryIsSanctionedResponse{IsSanctioned: false},
		},
		{
			name: "sanctioned address",
			req:  &sanction.QueryIsSanctionedRequest{Address: addrSanctioned.String()},
			exp:  &sanction.QueryIsSanctionedResponse{IsSanctioned: true},
		},
		{
			name: "temporarily sanctioned address",
			req:  &sanction.QueryIsSanctionedRequest{Address: AddrTempSanct.String()},
			exp:  &sanction.QueryIsSanctionedResponse{IsSanctioned: true},
		},
		{
			name: "sanctioned address that is temporarily unsanctioned",
			req:  &sanction.QueryIsSanctionedRequest{Address: addrSanctTempUn.String()},
			exp:  &sanction.QueryIsSanctionedResponse{IsSanctioned: false},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var resp *sanction.QueryIsSanctionedResponse
			var err error
			testFunc := func() {
				resp, err = s.Keeper.IsSanctioned(s.StdlibCtx, tc.req)
			}
			s.Require().NotPanics(testFunc, "IsSanctioned")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "IsSanctioned error")
			s.Assert().Equal(tc.exp, resp, "IsSanctioned response")
		})
	}
}

func (s *QueryTestSuite) TestKeeper_SanctionedAddresses() {
	addr1 := sdk.AccAddress("1_addr_made_for_test")
	addr2 := sdk.AccAddress("2_addr_made_for_test")
	addr3 := sdk.AccAddress("3_addr_made_for_test")
	addr4 := sdk.AccAddress("4_addr_made_for_test")
	addr5 := sdk.AccAddress("5_addr_made_for_test")
	addr6 := sdk.AccAddress("6_addr_made_for_test")

	asNextKey := func(addr sdk.AccAddress) []byte {
		key := keeper.CreateSanctionedAddrKey(addr)
		return key[1:]
	}

	tests := []struct {
		name     string
		iniState *sanction.GenesisState
		req      *sanction.QuerySanctionedAddressesRequest
		exp      *sanction.QuerySanctionedAddressesResponse
		expErr   []string
	}{
		{
			name:     "nil req nothing to return",
			iniState: nil,
			req:      nil,
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: nil,
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "nil req stuff to return",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr3.String(), addr5.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr2, 1, true),
					newTempEntry(addr3, 1, false),
				},
			},
			req: nil,
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: []string{addr1.String(), addr3.String(), addr5.String()},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   3,
				},
			},
		},
		{
			name:     "empty req nothing to return",
			iniState: nil,
			req:      &sanction.QuerySanctionedAddressesRequest{},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: nil,
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "empty req stuff to return",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr3.String(), addr5.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr2, 1, true),
					newTempEntry(addr3, 1, false),
				},
			},
			req: &sanction.QuerySanctionedAddressesRequest{},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: []string{addr1.String(), addr3.String(), addr5.String()},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   3,
				},
			},
		},
		{
			name: "paginated by counts",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{
					addr1.String(), addr2.String(), addr3.String(),
					addr4.String(), addr5.String(), addr6.String(),
				},
			},
			req: &sanction.QuerySanctionedAddressesRequest{
				Pagination: &query.PageRequest{
					Offset:     2,
					Limit:      3,
					CountTotal: false,
					Reverse:    false,
				},
			},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: []string{addr3.String(), addr4.String(), addr5.String()},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr6),
					Total:   0,
				},
			},
		},
		{
			name: "paginated by counts reversed",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{
					addr1.String(), addr2.String(), addr3.String(),
					addr4.String(), addr5.String(), addr6.String(),
				},
			},
			req: &sanction.QuerySanctionedAddressesRequest{
				Pagination: &query.PageRequest{
					Offset:     2,
					Limit:      3,
					CountTotal: false,
					Reverse:    true,
				},
			},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: []string{addr4.String(), addr3.String(), addr2.String()},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr1),
					Total:   0,
				},
			},
		},
		{
			name: "paginated by key",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{
					addr1.String(), addr2.String(), addr3.String(),
					addr4.String(), addr5.String(), addr6.String(),
				},
			},
			req: &sanction.QuerySanctionedAddressesRequest{
				Pagination: &query.PageRequest{
					Key:        asNextKey(addr2),
					Limit:      2,
					CountTotal: false,
					Reverse:    false,
				},
			},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: []string{addr2.String(), addr3.String()},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr4),
					Total:   0,
				},
			},
		},
		{
			name: "paginated by key reversed",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{
					addr1.String(), addr2.String(), addr3.String(),
					addr4.String(), addr5.String(), addr6.String(),
				},
			},
			req: &sanction.QuerySanctionedAddressesRequest{
				Pagination: &query.PageRequest{
					Key:        asNextKey(addr4),
					Limit:      2,
					CountTotal: false,
					Reverse:    true,
				},
			},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: []string{addr4.String(), addr3.String()},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr2),
					Total:   0,
				},
			},
		},
		{
			name: "paginated count total",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{
					addr1.String(), addr2.String(), addr3.String(),
					addr4.String(), addr5.String(), addr6.String(),
				},
			},
			req: &sanction.QuerySanctionedAddressesRequest{
				Pagination: &query.PageRequest{
					Limit:      1,
					CountTotal: true,
				},
			},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: []string{addr1.String()},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr2),
					Total:   6,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ClearState()
			if tc.iniState != nil {
				s.Require().NotPanics(func() {
					s.Keeper.InitGenesis(s.SdkCtx, tc.iniState)
				}, "InitGenesis")
			}

			var resp *sanction.QuerySanctionedAddressesResponse
			var err error
			testFunc := func() {
				resp, err = s.Keeper.SanctionedAddresses(s.StdlibCtx, tc.req)
			}
			s.Require().NotPanics(testFunc, "SanctionedAddresses")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "SanctionedAddresses error")
			if !s.Assert().Equal(tc.exp, resp, "SanctionedAddresses response") && tc.exp != nil && resp != nil {
				s.Assert().Equal(tc.exp.Addresses, resp.Addresses, "SanctionedAddresses response Addresses")
				s.Assert().Equal(tc.exp.Pagination.NextKey, resp.Pagination.NextKey, "TemporaryEntries response Pagination.NextKey")
				s.Assert().Equal(tc.exp.Pagination.Total, resp.Pagination.Total, "TemporaryEntries response Pagination.Total")
			}
		})
	}
}

func (s *QueryTestSuite) TestKeeper_TemporaryEntries() {
	addr1 := sdk.AccAddress("1_addr_made_for_test")
	addr2 := sdk.AccAddress("2_addr_made_for_test")
	addr3 := sdk.AccAddress("3_addr_made_for_test")
	addr4 := sdk.AccAddress("4_addr_made_for_test")
	addr5 := sdk.AccAddress("5_addr_made_for_test")
	addr6 := sdk.AccAddress("6_addr_made_for_test")

	asNextKey := func(addr sdk.AccAddress, govPropID uint64) []byte {
		key := keeper.CreateTemporaryKey(addr, govPropID)
		return key[1:]
	}
	propNextKey := func(govPropID uint64) []byte {
		return sdk.Uint64ToBigEndian(govPropID)
	}

	_, _, _, _, _, _, _ = addr1, addr2, addr3, addr4, addr5, addr6, asNextKey

	tests := []struct {
		name     string
		iniState *sanction.GenesisState
		req      *sanction.QueryTemporaryEntriesRequest
		exp      *sanction.QueryTemporaryEntriesResponse
		expErr   []string
	}{
		{
			name:     "nil req nothing to return",
			iniState: nil,
			req:      nil,
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: nil,
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "nil req stuff to return",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
				},
			},
			req: nil,
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   5,
				},
			},
		},
		{
			name:     "empty req nothing to return",
			iniState: nil,
			req:      &sanction.QueryTemporaryEntriesRequest{},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: nil,
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "empty req stuff to return",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   5,
				},
			},
		},
		{
			name: "bad address given",
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: "badbadbad",
			},
			expErr: []string{"InvalidArgument", "invalid address", "decoding bech32 failed"},
		},
		{
			name: "paginated by count",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: "",
				Pagination: &query.PageRequest{
					Offset:     2,
					Limit:      3,
					CountTotal: false,
					Reverse:    false,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
				},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr5, 10),
					Total:   0,
				},
			},
		},
		{
			name: "paginated by count reversed",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: "",
				Pagination: &query.PageRequest{
					Offset:     2,
					Limit:      3,
					CountTotal: false,
					Reverse:    true,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 10, false),
				},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr4, 4),
					Total:   0,
				},
			},
		},
		{
			name: "paginated by key",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: "",
				Pagination: &query.PageRequest{
					Key:        asNextKey(addr3, 3),
					Limit:      4,
					CountTotal: false,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
				},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr5, 12),
					Total:   0,
				},
			},
		},
		{
			name: "paginated by key reversed",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: "",
				Pagination: &query.PageRequest{
					Key:        asNextKey(addr5, 11),
					Limit:      4,
					CountTotal: false,
					Reverse:    true,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr3, 3, true),
				},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr2, 2),
					Total:   0,
				},
			},
		},
		{
			name: "paginated count total",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: "",
				Pagination: &query.PageRequest{
					Limit:      1,
					CountTotal: true,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
				},
				Pagination: &query.PageResponse{
					NextKey: asNextKey(addr2, 1),
					Total:   10,
				},
			},
		},
		{
			name: "with address paginated by count",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: addr5.String(),
				Pagination: &query.PageRequest{
					Offset:     1,
					Limit:      2,
					CountTotal: false,
					Reverse:    false,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
				},
				Pagination: &query.PageResponse{
					NextKey: propNextKey(13),
					Total:   0,
				},
			},
		},
		{
			name: "with address paginated by count reversed",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: addr5.String(),
				Pagination: &query.PageRequest{
					Offset:     1,
					Limit:      2,
					CountTotal: false,
					Reverse:    true,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 11, true),
				},
				Pagination: &query.PageResponse{
					NextKey: propNextKey(10),
					Total:   0,
				},
			},
		},
		{
			name: "with address paginated by key",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: addr5.String(),
				Pagination: &query.PageRequest{
					Key:        propNextKey(12),
					Limit:      2,
					CountTotal: false,
					Reverse:    false,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "with address paginated by key reversed",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: addr5.String(),
				Pagination: &query.PageRequest{
					Key:        propNextKey(12),
					Limit:      2,
					CountTotal: false,
					Reverse:    true,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 11, true),
				},
				Pagination: &query.PageResponse{
					NextKey: propNextKey(10),
					Total:   0,
				},
			},
		},
		{
			name: "with address paginated count total",
			iniState: &sanction.GenesisState{
				SanctionedAddresses: []string{addr1.String(), addr4.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, false),
					newTempEntry(addr2, 1, true),
					newTempEntry(addr2, 2, false),
					newTempEntry(addr3, 3, true),
					newTempEntry(addr4, 4, false),
					newTempEntry(addr5, 10, false),
					newTempEntry(addr5, 11, true),
					newTempEntry(addr5, 12, false),
					newTempEntry(addr5, 13, false),
					newTempEntry(addr6, 2, true),
				},
			},
			req: &sanction.QueryTemporaryEntriesRequest{
				Address: addr5.String(),
				Pagination: &query.PageRequest{
					Limit:      1,
					CountTotal: true,
				},
			},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					newTempEntry(addr5, 10, false),
				},
				Pagination: &query.PageResponse{
					NextKey: propNextKey(11),
					Total:   4,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ClearState()
			if tc.iniState != nil {
				s.Require().NotPanics(func() {
					s.Keeper.InitGenesis(s.SdkCtx, tc.iniState)
				}, "InitGenesis")
			}

			var resp *sanction.QueryTemporaryEntriesResponse
			var err error
			testFunc := func() {
				resp, err = s.Keeper.TemporaryEntries(s.StdlibCtx, tc.req)
			}
			s.Require().NotPanics(testFunc, "TemporaryEntries")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "TemporaryEntries error")
			if !s.Assert().Equal(tc.exp, resp, "TemporaryEntries response") && tc.exp != nil && resp != nil {
				s.Assert().Equal(tc.exp.Entries, resp.Entries, "TemporaryEntries response Entries")
				s.Assert().Equal(tc.exp.Pagination.NextKey, resp.Pagination.NextKey, "TemporaryEntries response Pagination.NextKey")
				s.Assert().Equal(tc.exp.Pagination.Total, resp.Pagination.Total, "TemporaryEntries response Pagination.Total")
			}
		})
	}
}

func (s *QueryTestSuite) TestKeeper_Params() {
	origMinSanct := sanction.DefaultImmediateSanctionMinDeposit
	origMinUnsanct := sanction.DefaultImmediateUnsanctionMinDeposit
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origMinSanct
		sanction.DefaultImmediateUnsanctionMinDeposit = origMinUnsanct
	}()
	sanction.DefaultImmediateSanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("stestcoin", 5))
	sanction.DefaultImmediateUnsanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("utestcoin", 7))

	tests := []struct {
		name     string
		iniState *sanction.GenesisState
		req      *sanction.QueryParamsRequest
		exp      *sanction.QueryParamsResponse
	}{
		{
			name:     "nil req nothing in state",
			iniState: nil,
			req:      nil,
			exp:      &sanction.QueryParamsResponse{Params: sanction.DefaultParams()},
		},
		{
			name:     "empty req nothing in state",
			iniState: nil,
			req:      &sanction.QueryParamsRequest{},
			exp:      &sanction.QueryParamsResponse{Params: sanction.DefaultParams()},
		},
		{
			name: "nil req params set to nothing",
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   nil,
					ImmediateUnsanctionMinDeposit: nil,
				},
			},
			req: nil,
			exp: &sanction.QueryParamsResponse{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   nil,
					ImmediateUnsanctionMinDeposit: nil,
				},
			},
		},
		{
			name: "empty req params set to nothing",
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   nil,
					ImmediateUnsanctionMinDeposit: nil,
				},
			},
			req: &sanction.QueryParamsRequest{},
			exp: &sanction.QueryParamsResponse{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   nil,
					ImmediateUnsanctionMinDeposit: nil,
				},
			},
		},
		{
			name: "nil req params set with values",
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("newscoin", 55)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("newucoin", 77)),
				},
			},
			req: nil,
			exp: &sanction.QueryParamsResponse{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("newscoin", 55)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("newucoin", 77)),
				},
			},
		},
		{
			name: "empty req params set with values",
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("newscoin", 55)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("newucoin", 77)),
				},
			},
			req: &sanction.QueryParamsRequest{},
			exp: &sanction.QueryParamsResponse{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("newscoin", 55)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("newucoin", 77)),
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ClearState()
			if tc.iniState != nil {
				s.Require().NotPanics(func() {
					s.Keeper.InitGenesis(s.SdkCtx, tc.iniState)
				}, "InitGenesis")
			}

			var resp *sanction.QueryParamsResponse
			s.RequireNotPanicsNoError(func() error {
				var err error
				resp, err = s.Keeper.Params(s.StdlibCtx, tc.req)
				return err
			}, "Params")

			if !s.Assert().Equal(tc.exp, resp, "Params response") && tc.exp != nil && resp != nil {
				s.Assert().Equal(tc.exp.Params.ImmediateSanctionMinDeposit.String(),
					resp.Params.ImmediateSanctionMinDeposit.String(),
					"Params response Params.ImmediateSanctionMinDeposit")
				s.Assert().Equal(tc.exp.Params.ImmediateUnsanctionMinDeposit.String(),
					resp.Params.ImmediateUnsanctionMinDeposit.String(),
					"Params response Params.ImmediateUnsanctionMinDeposit")
			}
		})
	}
}
