package grpcgateway

import (
	"net/url"
	"regexp"
	"strings"
)

const maxBodySize = 1 << 20 // 1 MB

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
func matchURL(u *url.URL, regexpToQueryMetadata map[*regexp.Regexp]queryMetadata) *uriMatch {
	uriPath := strings.TrimRight(u.Path, "/")
	params := make(map[string]string)

	for reg, qmd := range regexpToQueryMetadata {
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
				params[name] = matches[i+1]
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
