package grpcgateway

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
)

func Test_createRegexMapping(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		wantWarn    bool
	}{
		{
			name: "no annotations should not warn",
		},
		{
			name: "different annotations should not warn",
			annotations: map[string]string{
				"/foo/bar/{baz}":     "",
				"/crypto/{currency}": "",
			},
		},
		{
			name: "duplicate annotations should warn",
			annotations: map[string]string{
				"/hello/{world}":      "",
				"/hello/{developers}": "",
			},
			wantWarn: true,
		},
	}
	buf := bytes.NewBuffer(nil)
	logger := log.NewLogger(buf)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createRegexMapping(logger, tt.annotations)
			if tt.wantWarn {
				require.NotEmpty(t, buf.String())
			} else {
				require.Empty(t, buf.String())
			}
		})
	}
}
