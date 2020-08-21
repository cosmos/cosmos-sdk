package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAppendEvents(t *testing.T) {
	e1 := NewEvent("transfer", NewAttribute("sender", "foo"))
	e2 := NewEvent("transfer", NewAttribute("sender", "bar"))
	a := Events{e1}
	b := Events{e2}
	c := a.AppendEvents(b)
	require.Equal(t, c, Events{e1, e2})
	require.Equal(t, c, Events{e1}.AppendEvent(NewEvent("transfer", NewAttribute("sender", "bar"))))
	require.Equal(t, c, Events{e1}.AppendEvents(Events{e2}))
}

func TestAppendAttributes(t *testing.T) {
	e := NewEvent("transfer", NewAttribute("sender", "foo"))
	e = e.AppendAttributes(NewAttribute("recipient", "bar"))
	require.Len(t, e.Attributes, 2)
	require.Equal(t, e, NewEvent("transfer", NewAttribute("sender", "foo"), NewAttribute("recipient", "bar")))
}

func TestEmptyEvents(t *testing.T) {
	require.Equal(t, EmptyEvents(), Events{})
}

func TestAttributeString(t *testing.T) {
	require.Equal(t, "foo: bar", NewAttribute("foo", "bar").String())
}

func TestToABCIEvents(t *testing.T) {
	e := Events{NewEvent("transfer", NewAttribute("sender", "foo"))}
	abciEvents := e.ToABCIEvents()
	require.Len(t, abciEvents, 1)
	require.Equal(t, abciEvents[0].Type, e[0].Type)
	require.Equal(t, abciEvents[0].Attributes, e[0].Attributes)
}

func TestEventManager(t *testing.T) {
	em := NewEventManager()
	event := NewEvent("reward", NewAttribute("x", "y"))
	events := Events{NewEvent("transfer", NewAttribute("sender", "foo"))}

	em.EmitEvents(events)
	em.EmitEvent(event)

	require.Len(t, em.Events(), 2)
	require.Equal(t, em.Events(), events.AppendEvent(event))
}

func TestStringifyEvents(t *testing.T) {
	e := Events{
		NewEvent("message", NewAttribute("sender", "foo")),
		NewEvent("message", NewAttribute("module", "bank")),
	}
	se := StringifyEvents(e.ToABCIEvents())

	expectedTxtStr := "\t\t- message\n\t\t\t- sender: foo\n\t\t\t- module: bank"
	require.Equal(t, expectedTxtStr, se.String())

	bz, err := json.Marshal(se)
	require.NoError(t, err)

	expectedJSONStr := "[{\"type\":\"message\",\"attributes\":[{\"key\":\"sender\",\"value\":\"foo\"},{\"key\":\"module\",\"value\":\"bank\"}]}]"
	require.Equal(t, expectedJSONStr, string(bz))
}

func TestMarkEventsToIndex(t *testing.T) {
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
			events:   events,
			expected: events,
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
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, MarkEventsToIndex(tc.events, tc.indexSet))
		})
	}
}
