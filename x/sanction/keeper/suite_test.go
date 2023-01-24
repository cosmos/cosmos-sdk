package keeper_test

import (
	"context"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

// newTempEntry creates a TemporaryEntry from iterator callback args.
func newTempEntry(addr sdk.AccAddress, govPropId uint64, isSanctioned bool) *sanction.TemporaryEntry {
	status := sanction.TEMP_STATUS_SANCTIONED
	if !isSanctioned {
		status = sanction.TEMP_STATUS_UNSANCTIONED
	}
	return &sanction.TemporaryEntry{
		Address:    addr.String(),
		ProposalId: govPropId,
		Status:     status,
	}
}

// newIndTempEntry creates a TemporaryEntry to represent a proposal index temporary entry.
func newIndTempEntry(govPropId uint64, addr sdk.AccAddress) *sanction.TemporaryEntry {
	return &sanction.TemporaryEntry{
		Address:    addr.String(),
		ProposalId: govPropId,
		Status:     sanction.TEMP_STATUS_UNSPECIFIED,
	}
}

// BaseTestSuite is a base suite.Suite with values and functions commonly needed for these keeper tests.
type BaseTestSuite struct {
	suite.Suite

	App       *simapp.SimApp
	SdkCtx    sdk.Context
	StdlibCtx context.Context
	Keeper    keeper.Keeper
	GovKeeper *MockGovKeeper

	BlockTime time.Time
}

func (s *BaseTestSuite) BaseSetup() {
	s.BlockTime = tmtime.Now()
	s.App = simapp.Setup(s.T(), false)
	s.SdkCtx = s.App.BaseApp.NewContext(false, tmproto.Header{}).WithBlockHeader(tmproto.Header{Time: s.BlockTime})
	s.StdlibCtx = sdk.WrapSDKContext(s.SdkCtx)
	s.GovKeeper = NewMockGovKeeper()
	s.Keeper = s.App.SanctionKeeper.OnlyTestsWithGovKeeper(s.GovKeeper)
}

// GetStore gets the sanction module's store.
func (s *BaseTestSuite) GetStore() sdk.KVStore {
	return s.SdkCtx.KVStore(s.Keeper.OnlyTestsGetStoreKey())
}

// ClearState deletes all entries in the sanction store.
func (s *BaseTestSuite) ClearState() {
	var keysToDelete [][]byte
	store := s.GetStore()
	iter := store.Iterator(nil, nil)
	closeIter := func() {
		if iter != nil {
			s.Require().NoError(iter.Close(), "iter.Close")
			iter = nil
		}
	}
	defer closeIter()

	for ; iter.Valid(); iter.Next() {
		keysToDelete = append(keysToDelete, iter.Key())
	}
	closeIter()

	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// GetAllTempEntries gets all temporary entries in the store.
func (s *BaseTestSuite) GetAllTempEntries() []*sanction.TemporaryEntry {
	var tempEntries []*sanction.TemporaryEntry
	tempCB := func(cbAddr sdk.AccAddress, cbGovPropId uint64, cbIsSanction bool) bool {
		tempEntries = append(tempEntries, newTempEntry(cbAddr, cbGovPropId, cbIsSanction))
		return false
	}
	s.Require().NotPanics(func() {
		s.Keeper.IterateTemporaryEntries(s.SdkCtx, nil, tempCB)
	}, "IterateTemporaryEntries")
	return tempEntries
}

// GetAllIndexTempEntries gets all the gov prop index temporary entries in the store.
func (s *BaseTestSuite) GetAllIndexTempEntries() []*sanction.TemporaryEntry {
	var tempIndEntries []*sanction.TemporaryEntry
	tempIndCB := func(cbGovPropId uint64, cbAddr sdk.AccAddress) bool {
		tempIndEntries = append(tempIndEntries, newIndTempEntry(cbGovPropId, cbAddr))
		return false
	}
	s.Require().NotPanics(func() {
		s.Keeper.IterateProposalIndexEntries(s.SdkCtx, nil, tempIndCB)
	}, "IterateProposalIndexEntries")
	return tempIndEntries
}

// RequireNotPanicsNoError calls RequireNotPanicsNoError with this suite's t.
func (s *BaseTestSuite) RequireNotPanicsNoError(f func() error, msgAndArgs ...interface{}) {
	s.T().Helper()
	testutil.RequireNotPanicsNoError(s.T(), f, msgAndArgs...)
}

// ReqOKSetParams calls SetParams, making sure it doesn't panic and doesn't return an error.
func (s *BaseTestSuite) ReqOKSetParams(params *sanction.Params) {
	s.T().Helper()
	s.RequireNotPanicsNoError(func() error {
		return s.Keeper.SetParams(s.SdkCtx, params)
	}, "SetParams")
}

// ReqOKAddPermSanct calls SanctionAddresses, making sure it doesn't panic and doesn't return an error.
func (s *BaseTestSuite) ReqOKAddPermSanct(addrArgNames string, addrs ...sdk.AccAddress) {
	s.T().Helper()
	s.RequireNotPanicsNoError(func() error {
		return s.Keeper.SanctionAddresses(s.SdkCtx, addrs...)
	}, "SanctionAddresses(%s)", addrArgNames)
}

// ReqOKAddPermUnsanct calls UnsanctionAddresses, making sure it doesn't panic and doesn't return an error.
func (s *BaseTestSuite) ReqOKAddPermUnsanct(addrArgNames string, addrs ...sdk.AccAddress) {
	s.T().Helper()
	s.RequireNotPanicsNoError(func() error {
		return s.Keeper.UnsanctionAddresses(s.SdkCtx, addrs...)
	}, "UnsanctionAddresses(%s)", addrArgNames)
}

// ReqOKAddTempSanct calls AddTemporarySanction, making sure it doesn't panic and doesn't return an error.
func (s *BaseTestSuite) ReqOKAddTempSanct(id uint64, addrArgNames string, addrs ...sdk.AccAddress) {
	s.T().Helper()
	s.RequireNotPanicsNoError(func() error {
		return s.Keeper.AddTemporarySanction(s.SdkCtx, id, addrs...)
	}, "AddTemporarySanction(%d, %s)", id, addrArgNames)
}

// ReqOKAddTempUnsanct calls AddTemporaryUnsanction, making sure it doesn't panic and doesn't return an error.
func (s *BaseTestSuite) ReqOKAddTempUnsanct(id uint64, addrArgNames string, addrs ...sdk.AccAddress) {
	s.T().Helper()
	s.RequireNotPanicsNoError(func() error {
		return s.Keeper.AddTemporaryUnsanction(s.SdkCtx, id, addrs...)
	}, "AddTemporaryUnsanction(%d, %s)", id, addrArgNames)
}

// ReqOKDelAddrTemp calls DeleteAddrTempEntries, making sure it doesn't panic.
func (s *BaseTestSuite) ReqOKDelAddrTemp(addrArgNames string, addrs ...sdk.AccAddress) {
	s.T().Helper()
	s.Require().NotPanics(func() {
		s.Keeper.DeleteAddrTempEntries(s.SdkCtx, addrs...)
	}, "DeleteAddrTempEntries(%s)", addrArgNames)
}

// ReqOKDelPropTemp calls DeleteGovPropTempEntries, making sure it doesn't panic.
func (s *BaseTestSuite) ReqOKDelPropTemp(id uint64) {
	s.Require().NotPanics(func() {
		s.Keeper.DeleteGovPropTempEntries(s.SdkCtx, id)
	}, "DeleteGovPropTempEntries(%d)", id)
}

// NewAny calls codectypes.NewAnyWithValue requiring it to not error.
func (s *BaseTestSuite) NewAny(value proto.Message) *codectypes.Any {
	rv, err := codectypes.NewAnyWithValue(value)
	s.Require().NoError(err, "NewAnyWithValue on a %T", value)
	return rv
}

// CustomAny calls codectypes.NewAnyWithValue requiring it to not error.
// Then it overwrites the resulting TypeUrl using the provided typeValue.
func (s *BaseTestSuite) CustomAny(typeValue proto.Message, value proto.Message) *codectypes.Any {
	rv, err := codectypes.NewAnyWithValue(value)
	s.Require().NoError(err, "NewAnyWithValue on a %T", value)
	if typeValue == nil {
		rv.TypeUrl = ""
	} else {
		rv.TypeUrl = "/" + proto.MessageName(typeValue)
	}
	return rv
}

// ExportAndCheck calls ExportGenesis, making sure it doesn't panic,
// then makes sure the result is as expected.
// Returns true if everything is okay.
func (s *BaseTestSuite) ExportAndCheck(expected *sanction.GenesisState) bool {
	s.T().Helper()

	var actual *sanction.GenesisState
	testFunc := func() {
		actual = s.Keeper.ExportGenesis(s.SdkCtx)
	}
	s.Require().NotPanics(testFunc, "ExportGenesis")
	if s.Assert().Equal(expected, actual, "ExportGenesis result") {
		return true
	}
	if expected != nil && actual != nil {
		// Okay, it failed, make assertions on each field to hopefully help highlight the differences.
		if !s.Assert().Equal(expected.Params, actual.Params, "ExportGenesis result Params") && expected.Params != nil && actual.Params != nil {
			s.Assert().Equal(expected.Params.ImmediateSanctionMinDeposit.String(),
				actual.Params.ImmediateSanctionMinDeposit.String(),
				"ExportGenesis result Params.ImmediateSanctionMinDeposit")
			s.Assert().Equal(expected.Params.ImmediateUnsanctionMinDeposit.String(),
				actual.Params.ImmediateUnsanctionMinDeposit.String(),
				"ExportGenesis result Params.ImmediateUnsanctionMinDeposit")
		}
		s.Assert().Equal(expected.SanctionedAddresses,
			actual.SanctionedAddresses,
			"ExportGenesis result SanctionedAddresses")
		s.Assert().Equal(expected.TemporaryEntries,
			actual.TemporaryEntries,
			"ExportGenesis result TemporaryEntries")
	}
	return false
}
