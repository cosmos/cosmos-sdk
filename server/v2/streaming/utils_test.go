package streaming

import (
	"testing"

	"cosmossdk.io/core/event"
	"github.com/stretchr/testify/require"
)

func TestIntoStreamingEvents(t *testing.T) {
	tests := []struct {
		name      string
		events    []event.Event
		expected  []*Event
		expectErr bool
	}{
		{
			name: "convert single event with attributes",
			events: []event.Event{
				event.NewEvent(
					"transfer",
					event.NewAttribute("sender", "addr1"),
					event.NewAttribute("recipient", "addr2"),
				),
			},
			expected: []*Event{
				{
					Type: "transfer",
					Attributes: []*EventAttribute{
						{Key: "sender", Value: "addr1"},
						{Key: "recipient", Value: "addr2"},
					},
				},
			},
			expectErr: false,
		},
		{
			name:      "convert empty event list",
			events:    []event.Event{},
			expected:  []*Event{},
			expectErr: false,
		},
		{
			name: "convert multiple events",
			events: []event.Event{
				event.NewEvent(
					"transfer",
					event.NewAttribute("amount", "100"),
				),
				event.NewEvent(
					"message",
					event.NewAttribute("module", "bank"),
				),
			},
			expected: []*Event{
				{
					Type: "transfer",
					Attributes: []*EventAttribute{
						{Key: "amount", Value: "100"},
					},
				},
				{
					Type: "message",
					Attributes: []*EventAttribute{
						{Key: "module", Value: "bank"},
					},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IntoStreamingEvents(tt.events)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tt.expected), len(result))

			for i, expectedEvent := range tt.expected {
				require.Equal(t, expectedEvent.Type, result[i].Type)
				require.Equal(t, len(expectedEvent.Attributes), len(result[i].Attributes))

				for j, attr := range expectedEvent.Attributes {
					require.Equal(t, attr.Key, result[i].Attributes[j].Key)
					require.Equal(t, attr.Value, result[i].Attributes[j].Value)
				}
			}
		})
	}
}
