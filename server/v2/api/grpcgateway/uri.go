package grpcgateway

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/cosmos/gogoproto/jsonpb"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/mitchellh/mapstructure"
)

const MaxBodySize = 1 << 20 // 1 MB

// URIMatch contains the matching results
type URIMatch struct {
	// QueryInputName is the fully qualified name of the proto input type of the query.
	QueryInputName string

	// Params are any wildcard params found in the request.
	//
	// example: foo/bar/{baz} - foo/bar/qux -> {baz: qux}
	Params map[string]string
}

func (uri URIMatch) HasParams() bool {
	return len(uri.Params) > 0
}

// matchURI checks if a given URI matches any pattern and extracts wildcard values
func matchURI(uri string, getPatternToQueryInputName map[string]string) *URIMatch {
	// Remove trailing slash if present
	uri = strings.TrimRight(uri, "/")

	// for simple cases, where there are no wildcards, we can just do a map lookup.
	if inputName, ok := getPatternToQueryInputName[uri]; ok {
		return &URIMatch{
			QueryInputName: inputName,
		}
	}

	for getPattern, queryInputName := range getPatternToQueryInputName {
		// Remove trailing slash from getPattern if present
		getPattern = strings.TrimRight(getPattern, "/")

		// Get regex getPattern and param names
		regexPattern, paramNames := patternToRegex(getPattern)

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
				QueryInputName: queryInputName,
				Params:         params,
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

func createMessageFromJSON(match *URIMatch, r *http.Request) (gogoproto.Message, error) {
	requestType := gogoproto.MessageType(match.QueryInputName)
	if requestType == nil {
		return nil, fmt.Errorf("unknown request type")
	}

	msg, ok := reflect.New(requestType.Elem()).Interface().(gogoproto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to create message instance")
	}

	defer r.Body.Close()
	limitedReader := io.LimitReader(r.Body, MaxBodySize)
	err := jsonpb.Unmarshal(limitedReader, msg)
	if err != nil {
		return nil, fmt.Errorf("error parsing body: %w", err)
	}

	return msg, nil

}

func createMessage(match *URIMatch) (gogoproto.Message, error) {
	requestType := gogoproto.MessageType(match.QueryInputName)
	if requestType == nil {
		return nil, fmt.Errorf("unknown request type")
	}

	msg, ok := reflect.New(requestType.Elem()).Interface().(gogoproto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to create message instance")
	}

	if match.HasParams() {
		// Create a map with the proper field names from protobuf tags
		fieldMap := make(map[string]string)
		v := reflect.ValueOf(msg).Elem()
		t := v.Type()

		for key, value := range match.Params {
			// Find the corresponding struct field
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				tag := field.Tag.Get("protobuf")
				if nameMatch := regexp.MustCompile(`name=(\w+)`).FindStringSubmatch(tag); len(nameMatch) > 1 {
					if nameMatch[1] == key {
						fieldMap[field.Name] = value // Use the actual field name
						break
					}
				}
			}
		}

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:           msg,
			WeaklyTypedInput: true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create decoder: %w", err)
		}

		if err := decoder.Decode(fieldMap); err != nil {
			return nil, fmt.Errorf("failed to decode params: %w", err)
		}
	}
	return msg, nil
}
