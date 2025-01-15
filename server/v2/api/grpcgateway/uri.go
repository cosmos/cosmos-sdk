package grpcgateway

import (
	"net/url"
	"regexp"
	"strings"
)

// uriMatcher provides functionality to match HTTP request URIs.
type uriMatcher struct {
	// wildcardURIMatchers are used for complex URIs that involve wildcards (i.e. /foo/{bar}/baz)
	wildcardURIMatchers map[*regexp.Regexp]queryMetadata
	// simpleMatchers are used for simple URI's that have no wildcards (i.e. /foo/bar/baz).
	simpleMatchers map[string]queryMetadata
}

// uriMatch contains information related to a URI match.
type uriMatch struct {
	// QueryInputName is the fully qualified name of the proto input type of the query rpc method.
	QueryInputName string

	// Params are any wildcard params found in the request.
	//
	// example: /foo/bar/{baz} -> /foo/bar/hello = {"baz": "hello"}
	Params map[string]string
}

// matchURL attempts to find a match for the given URL.
// NOTE: if no match is found, nil is returned.
func (m uriMatcher) matchURL(u *url.URL) *uriMatch {
	// the url.RawPath is non-empty when URL encoded values are detected in the path params.
	// this requires different handling when getting path parameter values.
	isURLEncoded := false
	var uriPath string
	if u.RawPath != "" {
		isURLEncoded = true
		uriPath = u.RawPath
	} else {
		uriPath = u.Path
	}
	uriPath = strings.TrimRight(uriPath, "/")
	params := make(map[string]string)

	//  see if we can get a simple match first.
	if qmd, ok := m.simpleMatchers[uriPath]; ok {
		return &uriMatch{
			QueryInputName: qmd.queryInputProtoName,
			Params:         params,
		}
	}

	// try the complex matchers.
	for reg, qmd := range m.wildcardURIMatchers {
		matches := reg.FindStringSubmatch(uriPath)
		switch {
		case len(matches) == 1:
			return &uriMatch{
				QueryInputName: qmd.queryInputProtoName,
				Params:         params,
			}
		case len(matches) > 1:
			// first match is the URI, subsequent matches are the wild card values.
			for i, name := range qmd.wildcardKeyNames {
				var matchValue string
				if isURLEncoded {
					// we'll try to unescape the URL encoded values,
					// but if that doesn't work, we can just try the raw value.
					if decodedMatch, err := url.QueryUnescape(matches[i+1]); err == nil {
						matchValue = matches[i+1]
					} else {
						matchValue = decodedMatch
					}
				} else {
					matchValue = matches[i+1]
				}
				params[name] = matchValue
			}

			return &uriMatch{
				QueryInputName: qmd.queryInputProtoName,
				Params:         params,
			}
		}
	}
	return nil
}

// patternToRegex converts a URI pattern with wildcards to a regex pattern.
// Returns the regex pattern and a slice of wildcard names in order
func patternToRegex(pattern string) (string, []string) {
	escaped := regexp.QuoteMeta(pattern)
	var wildcardNames []string

	// extract and replace {param=**} patterns
	r1 := regexp.MustCompile(`\\\{([^}]+?)=\\\*\\\*\\}`)
	escaped = r1.ReplaceAllStringFunc(escaped, func(match string) string {
		// extract wildcard name without the =** suffix
		name := regexp.MustCompile(`\\\{(.+?)=`).FindStringSubmatch(match)[1]
		wildcardNames = append(wildcardNames, name)
		return "(.+)"
	})

	// extract and replace {param} patterns
	r2 := regexp.MustCompile(`\\\{([^}]+)\\}`)
	escaped = r2.ReplaceAllStringFunc(escaped, func(match string) string {
		// extract wildcard name from the curl braces {}.
		name := regexp.MustCompile(`\\\{(.*?)\\}`).FindStringSubmatch(match)[1]
		wildcardNames = append(wildcardNames, name)
		return "([^/]+)"
	})

	return "^" + escaped + "$", wildcardNames
}
