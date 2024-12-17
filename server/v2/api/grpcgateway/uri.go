package grpcgateway

import (
	"regexp"
	"strings"
)

// URIMatch contains the matching results
type URIMatch struct {
	MethodName string
	Params     map[string]string
}

func (uri URIMatch) HasParams() bool {
	return len(uri.Params) > 0
}

// matchURI checks if a given URI matches any pattern and extracts wildcard values
func matchURI(uri string, patterns map[string]string) *URIMatch {
	// Remove trailing slash if present
	uri = strings.TrimRight(uri, "/")

	for pattern, methodName := range patterns {
		// Remove trailing slash from pattern if present
		pattern = strings.TrimRight(pattern, "/")

		// Get regex pattern and param names
		regexPattern, paramNames := patternToRegex(pattern)

		// Compile and match
		regex := regexp.MustCompile(regexPattern)
		matches := regex.FindStringSubmatch(uri)

		if matches != nil && len(matches) > 1 {
			// First match is the full string, subsequent matches are capture groups
			params := make(map[string]string)
			for i, name := range paramNames {
				params[name] = matches[i+1]
			}

			return &URIMatch{
				MethodName: methodName,
				Params:     params,
			}
		}
	}

	return nil
}

// patternToRegex converts a URI pattern with wildcards to a regex pattern
// Returns the regex pattern and a slice of parameter names in order
func patternToRegex(pattern string) (string, []string) {
	escaped := regexp.QuoteMeta(pattern)
	var paramNames []string

	// Extract and replace {param=**} patterns
	r1 := regexp.MustCompile(`\\\{([^}]+?)=\\\*\\\*\\\}`)
	escaped = r1.ReplaceAllStringFunc(escaped, func(match string) string {
		// Extract param name without the =** suffix
		name := regexp.MustCompile(`\\\{(.+?)=`).FindStringSubmatch(match)[1]
		paramNames = append(paramNames, name)
		return "(.+)"
	})

	// Extract and replace {param} patterns
	r2 := regexp.MustCompile(`\\\{([^}]+)\\\}`)
	escaped = r2.ReplaceAllStringFunc(escaped, func(match string) string {
		// Extract param name from between { and }
		name := regexp.MustCompile(`\\\{(.*?)\\\}`).FindStringSubmatch(match)[1]
		paramNames = append(paramNames, name)
		return "([^/]+)"
	})

	return "^" + escaped + "$", paramNames
}
