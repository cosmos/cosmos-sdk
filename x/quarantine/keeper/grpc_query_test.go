package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/quarantine"

	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

// These tests are initiated by TestKeeperTestSuite in keeper_test.go

func (s *TestSuite) TestIsQuarantined() {
	addrQuarantinedAcc := MakeTestAddr("iq", 0)
	addrQuarantinedStr := addrQuarantinedAcc.String()
	addrNormalStr := MakeTestAddr("iq", 1).String()

	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, addrQuarantinedAcc), "SetOptIn")

	tests := []struct {
		name string
		req  *quarantine.QueryIsQuarantinedRequest
		resp *quarantine.QueryIsQuarantinedResponse
		err  []string
	}{
		{
			name: "nil req",
			req:  nil,
			err:  []string{"empty request"},
		},
		{
			name: "empty to address",
			req: &quarantine.QueryIsQuarantinedRequest{
				ToAddress: "",
			},
			err: []string{"to address cannot be empty"},
		},
		{
			name: "bad to address",
			req: &quarantine.QueryIsQuarantinedRequest{
				ToAddress: "yupitisbad",
			},
			err: []string{"invalid to address", "decoding bech32 failed"},
		},
		{
			name: "quarantined address",
			req: &quarantine.QueryIsQuarantinedRequest{
				ToAddress: addrQuarantinedStr,
			},
			resp: &quarantine.QueryIsQuarantinedResponse{
				IsQuarantined: true,
			},
		},
		{
			name: "normal address",
			req: &quarantine.QueryIsQuarantinedRequest{
				ToAddress: addrNormalStr,
			},
			resp: &quarantine.QueryIsQuarantinedResponse{
				IsQuarantined: false,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := s.keeper.IsQuarantined(s.stdlibCtx, tc.req)
			if s.AssertErrorContents(err, tc.err, "IsQuarantined error") {
				s.Assert().Equal(tc.resp, resp, "IsQuarantined result")
			}
		})
	}
}

func (s *TestSuite) TestQuarantinedFunds() {
	makeAddr := func(index uint8) (sdk.AccAddress, string) {
		addr := MakeTestAddr("qf", index)
		return addr, addr.String()
	}
	newQF := func(coins string, toAddr string, fromAddrs ...string) *quarantine.QuarantinedFunds {
		return &quarantine.QuarantinedFunds{
			ToAddress:               toAddr,
			UnacceptedFromAddresses: fromAddrs,
			Coins:                   s.cz(coins),
			Declined:                false,
		}
	}
	qfz := func(qfs ...*quarantine.QuarantinedFunds) []*quarantine.QuarantinedFunds {
		return qfs
	}
	addr0Acc, addr0Str := makeAddr(0)
	addr1Acc, addr1Str := makeAddr(1)
	_, addr2Str := makeAddr(2)
	addr3Acc, addr3Str := makeAddr(3)
	addr4Acc, addr4Str := makeAddr(4)
	_, addr5Str := makeAddr(5)

	// 0 <- 1 = nothing
	qf023 := newQF("3rec", addr0Str, addr2Str, addr3Str)
	qf03 := newQF("7rec", addr0Str, addr3Str)
	qf23 := newQF("17goldcoin", addr2Str, addr3Str)
	// 1 <- anyone = nothing
	qf31 := newQF("37goldcoin", addr3Str, addr1Str)
	qf32 := newQF("79goldcoin", addr3Str, addr2Str)
	qf40 := newQF("163goldcoin", addr4Str, addr0Str)
	qf42d := newQF("331goldcoin", addr4Str, addr2Str)
	qf42d.Declined = true
	qf43 := newQF("673goldcoin", addr4Str, addr3Str)
	qf50 := newQF("1361goldcoin", addr5Str, addr0Str)
	qf51 := newQF("2729goldcoin", addr5Str, addr1Str)
	qf52 := newQF("5471goldcoin", addr5Str, addr2Str)
	qf53 := newQF("10949goldcoin", addr5Str, addr3Str)
	qf54 := newQF("21911goldcoin", addr5Str, addr4Str)
	qf513d := newQF("43853goldcoin", addr5Str, addr1Str, addr3Str)
	qf513d.Declined = true
	qf501234 := newQF("43853goldcoin", addr5Str, addr0Str, addr1Str, addr2Str, addr3Str, addr4Str)
	allQuarantinedFunds := qfz(
		qf023, qf03,
		qf23,
		qf31, qf32,
		qf40, qf42d, qf43,
		qf50, qf51, qf52, qf53, qf54, qf513d, qf501234,
	)

	for i, qf := range allQuarantinedFunds {
		toAddr, taerr := sdk.AccAddressFromBech32(qf.ToAddress)
		s.Require().NoError(taerr, "AccAddressFromBech32 allQuarantinedFunds[%d].ToAddress", i)
		var qr *quarantine.QuarantineRecord
		testFuncAsQR := func() {
			qr = quarantine.NewQuarantineRecord(qf.UnacceptedFromAddresses, qf.Coins, qf.Declined)
		}
		s.Require().NotPanics(testFuncAsQR, "NewQuarantineRecord allQuarantinedFunds[%d]", i)
		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, qr)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord allQuarantinedFunds[%d]", i)
	}

	tests := []struct {
		name string
		req  *quarantine.QueryQuarantinedFundsRequest
		resp *quarantine.QueryQuarantinedFundsResponse
		err  []string
	}{
		{
			name: "nil req",
			req:  nil,
			err:  []string{"empty request"},
		},
		{
			name: "from without to",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress:   "",
				FromAddress: addr0Str,
			},
			err: []string{"to address cannot be empty when from address is not"},
		},
		{
			name: "bad to address",
			req:  &quarantine.QueryQuarantinedFundsRequest{ToAddress: "baaaaaaaaad1sheep"},
			err:  []string{"invalid to address", "decoding bech32 failed"},
		},
		{
			name: "bad from address",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress:   addr0Str,
				FromAddress: "still1badsheep",
			},
			err: []string{"invalid from address", "decoding bech32 failed"},
		},
		{
			name: "to and from  has no entries",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress:   addr0Str,
				FromAddress: addr1Str,
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: nil,
				Pagination:       nil,
			},
		},
		{
			name: "to and from  has one entry",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress:   addr0Str,
				FromAddress: addr2Str,
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: qfz(qf023),
				Pagination:       nil,
			},
		},
		{
			name: "to and from  has two entries",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress:   addr0Str,
				FromAddress: addr3Str,
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: qfz(qf03, qf023),
				Pagination:       nil,
			},
		},
		{
			name: "only to  has no entries",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress: addr1Str,
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: nil,
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "only to has one entry",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress: addr2Str,
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: qfz(qf23),
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   1,
				},
			},
		},
		{
			name: "only to has two entries",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress: addr3Str,
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: qfz(qf31, qf32),
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   2,
				},
			},
		},
		{
			name: "only to does not include declined funds",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress: addr4Str,
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: qfz(qf40, qf43),
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   2,
				},
			},
		},
		{
			name: "only to with page req",
			req: &quarantine.QueryQuarantinedFundsRequest{
				ToAddress: addr5Str,
				Pagination: &query.PageRequest{
					Key:        address.MustLengthPrefix(addr1Acc),
					Offset:     0,
					Limit:      2,
					CountTotal: false,
					Reverse:    false,
				},
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: qfz(qf51, qf52),
				Pagination: &query.PageResponse{
					NextKey: address.MustLengthPrefix(addr3Acc),
					Total:   0,
				},
			},
		},
		{
			name: "get all",
			req:  &quarantine.QueryQuarantinedFundsRequest{},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: qfz(
					qf03, qf023,
					qf23,
					qf31, qf32,
					qf40, qf43,
					qf50, qf51, qf52, qf53, qf54, qf501234,
				),
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   13,
				},
			},
		},
		{
			name: "get all with page req",
			req: &quarantine.QueryQuarantinedFundsRequest{
				Pagination: &query.PageRequest{
					Key:        nil,
					Offset:     3,
					Limit:      3,
					CountTotal: true,
					Reverse:    false,
				},
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: qfz(qf31, qf32, qf40),
				Pagination: &query.PageResponse{
					NextKey: quarantine.CreateRecordKey(addr4Acc, addr3Acc)[1:],
					Total:   13,
				},
			},
		},
		{
			name: "get all with page rev",
			req: &quarantine.QueryQuarantinedFundsRequest{
				Pagination: &query.PageRequest{
					Key:        nil,
					Offset:     4,
					Limit:      3,
					CountTotal: true,
					Reverse:    true,
				},
			},
			resp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: qfz(qf51, qf50, qf43),
				Pagination: &query.PageResponse{
					NextKey: quarantine.CreateRecordKey(addr4Acc, addr0Acc)[1:],
					Total:   13,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := s.keeper.QuarantinedFunds(s.stdlibCtx, tc.req)
			if s.AssertErrorContents(err, tc.err, "QuarantinedFunds error") {
				s.Assert().Equal(tc.resp, resp, "QuarantinedFunds response")
			}
		})
	}
}

func (s *TestSuite) TestAutoResponses() {
	makeAddr := func(index uint8) (sdk.AccAddress, string) {
		addr := MakeTestAddr("ar", index)
		return addr, addr.String()
	}
	newARE := func(toAddr, fromAddr string, resp quarantine.AutoResponse) *quarantine.AutoResponseEntry {
		return &quarantine.AutoResponseEntry{
			ToAddress:   toAddr,
			FromAddress: fromAddr,
			Response:    resp,
		}
	}
	addr0Acc, addr0Str := makeAddr(0)
	addr1Acc, addr1Str := makeAddr(1)
	addr2Acc, addr2Str := makeAddr(2)
	addr3Acc, addr3Str := makeAddr(3)
	addr4Acc, _ := makeAddr(4)
	addr5Acc, addr5Str := makeAddr(5)
	addr6Acc, addr6Str := makeAddr(6)

	// Setup:
	// 0 <- 1 accept, 2 decline, 3 unspecified
	// 2 <- 3 accept
	// 3 <- 2 decline
	// 6 <- 0 decline, 1 accept, 2 decline, 3 accept, 4 unspecified, 5 accept
	s.keeper.SetAutoResponse(s.sdkCtx, addr0Acc, addr1Acc, quarantine.AUTO_RESPONSE_ACCEPT)
	s.keeper.SetAutoResponse(s.sdkCtx, addr0Acc, addr2Acc, quarantine.AUTO_RESPONSE_DECLINE)
	s.keeper.SetAutoResponse(s.sdkCtx, addr0Acc, addr3Acc, quarantine.AUTO_RESPONSE_UNSPECIFIED)
	s.keeper.SetAutoResponse(s.sdkCtx, addr2Acc, addr3Acc, quarantine.AUTO_RESPONSE_ACCEPT)
	s.keeper.SetAutoResponse(s.sdkCtx, addr3Acc, addr2Acc, quarantine.AUTO_RESPONSE_DECLINE)
	s.keeper.SetAutoResponse(s.sdkCtx, addr6Acc, addr0Acc, quarantine.AUTO_RESPONSE_DECLINE)
	s.keeper.SetAutoResponse(s.sdkCtx, addr6Acc, addr1Acc, quarantine.AUTO_RESPONSE_ACCEPT)
	s.keeper.SetAutoResponse(s.sdkCtx, addr6Acc, addr2Acc, quarantine.AUTO_RESPONSE_DECLINE)
	s.keeper.SetAutoResponse(s.sdkCtx, addr6Acc, addr3Acc, quarantine.AUTO_RESPONSE_ACCEPT)
	s.keeper.SetAutoResponse(s.sdkCtx, addr6Acc, addr4Acc, quarantine.AUTO_RESPONSE_UNSPECIFIED)
	s.keeper.SetAutoResponse(s.sdkCtx, addr6Acc, addr5Acc, quarantine.AUTO_RESPONSE_ACCEPT)

	tests := []struct {
		name string
		req  *quarantine.QueryAutoResponsesRequest
		resp *quarantine.QueryAutoResponsesResponse
		err  []string
	}{
		{
			name: "no req",
			req:  nil,
			err:  []string{"empty request"},
		},
		{
			name: "no to address",
			req: &quarantine.QueryAutoResponsesRequest{
				ToAddress:   "",
				FromAddress: addr1Str,
			},
			err: []string{"to address cannot be empty"},
		},
		{
			name: "bad to address",
			req: &quarantine.QueryAutoResponsesRequest{
				ToAddress:   "not1goodone",
				FromAddress: addr1Str,
			},
			err: []string{"invalid to address", "decoding bech32 failed"},
		},
		{
			name: "bad from address",
			req: &quarantine.QueryAutoResponsesRequest{
				ToAddress:   addr0Str,
				FromAddress: "also1badone",
			},
			err: []string{"invalid from address", "decoding bech32 failed"},
		},
		{
			name: "to and from auto-accept",
			req: &quarantine.QueryAutoResponsesRequest{
				ToAddress:   addr0Str,
				FromAddress: addr1Str,
			},
			resp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr0Str, addr1Str, quarantine.AUTO_RESPONSE_ACCEPT),
				},
				Pagination: nil,
			},
		},
		{
			name: "to and from auto-decline",
			req: &quarantine.QueryAutoResponsesRequest{
				ToAddress:   addr0Str,
				FromAddress: addr2Str,
			},
			resp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr0Str, addr2Str, quarantine.AUTO_RESPONSE_DECLINE),
				},
				Pagination: nil,
			},
		},
		{
			name: "to and from unspecified",
			req: &quarantine.QueryAutoResponsesRequest{
				ToAddress:   addr0Str,
				FromAddress: addr3Str,
			},
			resp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr0Str, addr3Str, quarantine.AUTO_RESPONSE_UNSPECIFIED),
				},
				Pagination: nil,
			},
		},
		{
			name: "to and same from",
			req: &quarantine.QueryAutoResponsesRequest{
				ToAddress:   addr0Str,
				FromAddress: addr0Str,
			},
			resp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr0Str, addr0Str, quarantine.AUTO_RESPONSE_ACCEPT),
				},
				Pagination: nil,
			},
		},
		{
			name: "only to with no entries",
			req:  &quarantine.QueryAutoResponsesRequest{ToAddress: addr1Str},
			resp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: nil,
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "only to with one entry accept",
			req:  &quarantine.QueryAutoResponsesRequest{ToAddress: addr2Str},
			resp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr2Str, addr3Str, quarantine.AUTO_RESPONSE_ACCEPT),
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   1,
				},
			},
		},
		{
			name: "only to with one entry decline",
			req:  &quarantine.QueryAutoResponsesRequest{ToAddress: addr3Str},
			resp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr3Str, addr2Str, quarantine.AUTO_RESPONSE_DECLINE),
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   1,
				},
			},
		},
		{
			name: "only to with two entries", // 0 = 1 2 (accept/decline)
			req:  &quarantine.QueryAutoResponsesRequest{ToAddress: addr0Str},
			resp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr0Str, addr1Str, quarantine.AUTO_RESPONSE_ACCEPT),
					newARE(addr0Str, addr2Str, quarantine.AUTO_RESPONSE_DECLINE),
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   2,
				},
			},
		},
		{
			name: "only to with page req",
			req: &quarantine.QueryAutoResponsesRequest{
				ToAddress: addr6Str,
				Pagination: &query.PageRequest{
					Key:        nil,
					Offset:     2,
					Limit:      4,
					CountTotal: true,
					Reverse:    false,
				},
			},
			resp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr6Str, addr2Str, quarantine.AUTO_RESPONSE_DECLINE),
					newARE(addr6Str, addr3Str, quarantine.AUTO_RESPONSE_ACCEPT),
					newARE(addr6Str, addr5Str, quarantine.AUTO_RESPONSE_ACCEPT),
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   5,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := s.keeper.AutoResponses(s.stdlibCtx, tc.req)
			if s.AssertErrorContents(err, tc.err, "AutoResponses error") {
				s.Assert().Equal(tc.resp, resp, "AutoResponses response")
			}
		})
	}
}
