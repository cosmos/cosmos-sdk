package grpcgateway

import (
	"net/url"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
)

func TestMatchURI(t *testing.T) {
	testCases := []struct {
		name     string
		uri      string
		mapping  map[string]string
		expected *uriMatch
	}{
		{
			name:     "simple match, no wildcards",
			uri:      "https://localhost:8080/foo/bar",
			mapping:  map[string]string{"/foo/bar": "query.Bank"},
			expected: &uriMatch{QueryInputName: "query.Bank", Params: map[string]string{}},
		},
		{
			name: "match with wildcard similar to simple match - simple",
			uri:  "https://localhost:8080/bank/supply/latest",
			mapping: map[string]string{
				"/bank/supply/{height}": "queryBankHeight",
				"/bank/supply/latest":   "queryBankLatest",
			},
			expected: &uriMatch{QueryInputName: "queryBankLatest", Params: map[string]string{}},
		},
		{
			name: "match with wildcard similar to simple match - wildcard",
			uri:  "https://localhost:8080/bank/supply/52",
			mapping: map[string]string{
				"/bank/supply/{height}": "queryBankHeight",
				"/bank/supply/latest":   "queryBankLatest",
			},
			expected: &uriMatch{QueryInputName: "queryBankHeight", Params: map[string]string{"height": "52"}},
		},
		{
			name:    "wildcard match at the end",
			uri:     "https://localhost:8080/foo/bar/buzz",
			mapping: map[string]string{"/foo/bar/{baz}": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string]string{"baz": "buzz"},
			},
		},
		{
			name:    "wildcard match in the middle",
			uri:     "https://localhost:8080/foo/buzz/bar",
			mapping: map[string]string{"/foo/{baz}/bar": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string]string{"baz": "buzz"},
			},
		},
		{
			name:    "multiple wild cards",
			uri:     "https://localhost:8080/foo/bar/baz/buzz",
			mapping: map[string]string{"/foo/bar/{q1}/{q2}": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string]string{"q1": "baz", "q2": "buzz"},
			},
		},
		{
			name:    "catch-all wildcard",
			uri:     "https://localhost:8080/foo/bar/ibc/token/stuff",
			mapping: map[string]string{"/foo/bar/{ibc_token=**}": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string]string{"ibc_token": "ibc/token/stuff"},
			},
		},
		{
			name:     "no match should return nil",
			uri:      "https://localhost:8080/foo/bar",
			mapping:  map[string]string{"/bar/foo": "bar"},
			expected: nil,
		},
	}

	logger := log.NewLogger(os.Stdout)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.uri)
			require.NoError(t, err)

			regexpMatchers, simpleMatchers := createRegexMapping(logger, tc.mapping)
			matcher := uriMatcher{
				wildcardURIMatchers: regexpMatchers,
				simpleMatchers:      simpleMatchers,
			}

			actual := matcher.matchURL(u)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_patternToRegex(t *testing.T) {
	tests := []struct {
		name           string
		pattern        string
		wildcards      []string
		wildcardValues []string
		shouldMatch    string
		shouldNotMatch []string
	}{
		{
			name:           "simple match, no wildcards",
			pattern:        "/foo/bar/baz",
			shouldMatch:    "/foo/bar/baz",
			shouldNotMatch: []string{"/foo/bar", "/foo", "/foo/bar/baz/boo"},
		},
		{
			name:           "match with wildcard",
			pattern:        "/foo/bar/{baz}",
			wildcards:      []string{"baz"},
			shouldMatch:    "/foo/bar/hello",
			wildcardValues: []string{"hello"},
			shouldNotMatch: []string{"/foo/bar", "/foo/bar/baz/boo"},
		},
		{
			name:           "match with multiple wildcards",
			pattern:        "/foo/{bar}/{baz}/meow",
			wildcards:      []string{"bar", "baz"},
			shouldMatch:    "/foo/hello/world/meow",
			wildcardValues: []string{"hello", "world"},
			shouldNotMatch: []string{"/foo/bar/baz/boo", "/foo/bar/baz"},
		},
		{
			name:           "match catch-all wildcard",
			pattern:        `/foo/bar/{baz=**}`,
			wildcards:      []string{"baz"},
			shouldMatch:    `/foo/bar/this/is/a/long/wildcard`,
			wildcardValues: []string{"this/is/a/long/wildcard"},
			shouldNotMatch: []string{"/foo/bar", "/foo", "/foo/baz/bar/long/wild/card"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regString, wildcards := patternToRegex(tt.pattern)
			// should produce the same wildcard keys
			require.Equal(t, tt.wildcards, wildcards)
			reg := regexp.MustCompile(regString)

			// handle the "should match" case.
			matches := reg.FindStringSubmatch(tt.shouldMatch)
			require.True(t, len(matches) > 0) // there should always be a match.
			// when matches > 1, this means we got wildcard values to handle. the test should have wildcard values.
			if len(matches) > 1 {
				require.Greater(t, len(tt.wildcardValues), 0)
			}
			// matches[0] is the URL, everything else should be those wildcard values.
			if len(tt.wildcardValues) > 0 {
				require.Equal(t, matches[1:], tt.wildcardValues)
			}

			// should never match these.
			for _, notMatch := range tt.shouldNotMatch {
				require.Len(t, reg.FindStringSubmatch(notMatch), 0)
			}
		})
	}
}
