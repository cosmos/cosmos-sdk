package types_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	testdata "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type eventsTestSuite struct {
	suite.Suite
}

func TestEventsTestSuite(t *testing.T) {
	suite.Run(t, new(eventsTestSuite))
}

func (s *eventsTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *eventsTestSuite) TestAppendEvents() {
	e1 := sdk.NewEvent("transfer", sdk.NewAttribute("sender", "foo"))
	e2 := sdk.NewEvent("transfer", sdk.NewAttribute("sender", "bar"))
	a := sdk.Events{e1}
	b := sdk.Events{e2}
	c := a.AppendEvents(b)
	s.Require().Equal(c, sdk.Events{e1, e2})
	s.Require().Equal(c, sdk.Events{e1}.AppendEvent(sdk.NewEvent("transfer", sdk.NewAttribute("sender", "bar"))))
	s.Require().Equal(c, sdk.Events{e1}.AppendEvents(sdk.Events{e2}))
}

func (s *eventsTestSuite) TestAppendAttributes() {
	e := sdk.NewEvent("transfer", sdk.NewAttribute("sender", "foo"))
	e = e.AppendAttributes(sdk.NewAttribute("recipient", "bar"))
	s.Require().Len(e.Attributes, 2)
	s.Require().Equal(e, sdk.NewEvent("transfer", sdk.NewAttribute("sender", "foo"), sdk.NewAttribute("recipient", "bar")))
}

func (s *eventsTestSuite) TestEmptyEvents() {
	s.Require().Equal(sdk.EmptyEvents(), sdk.Events{})
}

func (s *eventsTestSuite) TestAttributeString() {
	s.Require().Equal("foo: bar", sdk.NewAttribute("foo", "bar").String())
}

func (s *eventsTestSuite) TestToABCIEvents() {
	e := sdk.Events{sdk.NewEvent("transfer", sdk.NewAttribute("sender", "foo"))}
	abciEvents := e.ToABCIEvents()
	s.Require().Len(abciEvents, 1)
	s.Require().Equal(abciEvents[0].Type, e[0].Type)
	s.Require().Equal(abciEvents[0].Attributes, e[0].Attributes)
}

func (s *eventsTestSuite) TestEventManager() {
	em := sdk.NewEventManager()
	event := sdk.NewEvent("reward", sdk.NewAttribute("x", "y"))
	events := sdk.Events{sdk.NewEvent("transfer", sdk.NewAttribute("sender", "foo"))}

	em.EmitEvents(events)
	em.EmitEvent(event)

	s.Require().Len(em.Events(), 2)
	s.Require().Equal(em.Events(), events.AppendEvent(event))
}

func (s *eventsTestSuite) TestEventManagerTypedEvents() {
	em := sdk.NewEventManager()

	coin := sdk.NewCoin("fakedenom", sdk.NewInt(1999999))
	cat := testdata.Cat{
		Moniker: "Garfield",
		Lives:   6,
	}
	animal, err := codectypes.NewAnyWithValue(&cat)
	s.Require().NoError(err)
	hasAnimal := testdata.HasAnimal{
		X:      1000,
		Animal: animal,
	}

	s.Require().NoError(em.EmitTypedEvents(&coin))
	s.Require().NoError(em.EmitTypedEvent(&hasAnimal))
	s.Require().Len(em.Events(), 2)

	msg1, err := sdk.ParseTypedEvent(em.Events().ToABCIEvents()[0])
	s.Require().NoError(err)
	s.Require().Equal(coin.String(), msg1.String())
	s.Require().Equal(reflect.TypeOf(&coin), reflect.TypeOf(msg1))

	msg2, err := sdk.ParseTypedEvent(em.Events().ToABCIEvents()[1])
	s.Require().NoError(err)
	s.Require().Equal(reflect.TypeOf(&hasAnimal), reflect.TypeOf(msg2))
	response := msg2.(*testdata.HasAnimal)
	s.Require().Equal(hasAnimal.Animal.String(), response.Animal.String())
}

func (s *eventsTestSuite) TestStringifyEvents() {
	e := sdk.Events{
		sdk.NewEvent("message", sdk.NewAttribute("sender", "foo")),
		sdk.NewEvent("message", sdk.NewAttribute("module", "bank")),
	}
	se := sdk.StringifyEvents(e.ToABCIEvents())

	expectedTxtStr := "\t\t- message\n\t\t\t- sender: foo\n\t\t\t- module: bank"
	s.Require().Equal(expectedTxtStr, se.String())

	bz, err := json.Marshal(se)
	s.Require().NoError(err)

	expectedJSONStr := "[{\"type\":\"message\",\"attributes\":[{\"key\":\"sender\",\"value\":\"foo\"},{\"key\":\"module\",\"value\":\"bank\"}]}]"
	s.Require().Equal(expectedJSONStr, string(bz))
}

func (s *eventsTestSuite) TestMarkEventsToIndex() {
	events := []abci.Event{
		{
			Type: "message",
			Attributes: []abci.EventAttribute{
				{Key: []byte("sender"), Value: []byte("foo")},
				{Key: []byte("recipient"), Value: []byte("bar")},
			},
		},
		{
			Type: "staking",
			Attributes: []abci.EventAttribute{
				{Key: []byte("deposit"), Value: []byte("5")},
				{Key: []byte("unbond"), Value: []byte("10")},
			},
		},
	}

	testCases := map[string]struct {
		events   []abci.Event
		indexSet map[string]struct{}
		expected []abci.Event
	}{
		"empty index set": {
			events: events,
			expected: []abci.Event{
				{
					Type: "message",
					Attributes: []abci.EventAttribute{
						{Key: []byte("sender"), Value: []byte("foo"), Index: true},
						{Key: []byte("recipient"), Value: []byte("bar"), Index: true},
					},
				},
				{
					Type: "staking",
					Attributes: []abci.EventAttribute{
						{Key: []byte("deposit"), Value: []byte("5"), Index: true},
						{Key: []byte("unbond"), Value: []byte("10"), Index: true},
					},
				},
			},
			indexSet: map[string]struct{}{},
		},
		"index some events": {
			events: events,
			expected: []abci.Event{
				{
					Type: "message",
					Attributes: []abci.EventAttribute{
						{Key: []byte("sender"), Value: []byte("foo"), Index: true},
						{Key: []byte("recipient"), Value: []byte("bar")},
					},
				},
				{
					Type: "staking",
					Attributes: []abci.EventAttribute{
						{Key: []byte("deposit"), Value: []byte("5"), Index: true},
						{Key: []byte("unbond"), Value: []byte("10")},
					},
				},
			},
			indexSet: map[string]struct{}{
				"message.sender":  {},
				"staking.deposit": {},
			},
		},
		"index all events": {
			events: events,
			expected: []abci.Event{
				{
					Type: "message",
					Attributes: []abci.EventAttribute{
						{Key: []byte("sender"), Value: []byte("foo"), Index: true},
						{Key: []byte("recipient"), Value: []byte("bar"), Index: true},
					},
				},
				{
					Type: "staking",
					Attributes: []abci.EventAttribute{
						{Key: []byte("deposit"), Value: []byte("5"), Index: true},
						{Key: []byte("unbond"), Value: []byte("10"), Index: true},
					},
				},
			},
			indexSet: map[string]struct{}{
				"message.sender":    {},
				"message.recipient": {},
				"staking.deposit":   {},
				"staking.unbond":    {},
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		s.T().Run(name, func(_ *testing.T) {
			s.Require().Equal(tc.expected, sdk.MarkEventsToIndex(tc.events, tc.indexSet))
		})
	}
}
