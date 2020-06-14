package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilRoute(t *testing.T) {
	tests := []struct {
		name     string
		route    Route
		expected bool
	}{
		{
			name:     "all empty",
			route:    NewRoute("", nil),
			expected: true,
		},
		{
			name:     "only path",
			route:    NewRoute("some", nil),
			expected: true,
		},
		{
			name: "only handler",
			route: NewRoute("", func(ctx Context, msg Msg) (*Result, error) {
				return nil, nil
			}),
			expected: true,
		},
		{
			name: "correct route",
			route: NewRoute("some", func(ctx Context, msg Msg) (*Result, error) {
				return nil, nil
			}),
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.route.Empty())
		})
	}
}
