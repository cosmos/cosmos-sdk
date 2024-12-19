package grpcgateway

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/cosmos/gogoproto/jsonpb"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxBodySize = 1 << 20 // 1 MB

// uriMatch contains information related to a URI match.
type uriMatch struct {
	// QueryInputName is the fully qualified name of the proto input type of the query rpc method.
	QueryInputName string

	// Params are any wildcard params found in the request.
	//
	// example: foo/bar/{baz} - foo/bar/qux -> {baz: qux}
	Params map[string]string
}

// HasParams reports whether the uriMatch has any params.
func (uri uriMatch) HasParams() bool {
	return len(uri.Params) > 0
}

// matchURL attempts to find a match for the given URL.
// NOTE: if no match is found, nil is returned.
func matchURL(u *url.URL, getPatternToQueryInputName map[string]string) *uriMatch {
	uriPath := strings.TrimRight(u.Path, "/")
	queryParams := u.Query()

	params := make(map[string]string)
	for key, vals := range queryParams {
		if len(vals) > 0 {
			// url.Values contains a slice for the values as you are able to specify a key multiple times in URL.
			// example: https://localhost:9090/do/something?color=red&color=blue&color=green
			// We will just take the first value in the slice.
			params[key] = vals[0]
		}
	}

	// for simple cases where there are no wildcards, we can just do a map lookup.
	if inputName, ok := getPatternToQueryInputName[uriPath]; ok {
		return &uriMatch{
			QueryInputName: inputName,
			Params:         params,
		}
	}

	// attempt to find a match in the pattern map.
	for getPattern, queryInputName := range getPatternToQueryInputName {
		getPattern = strings.TrimRight(getPattern, "/")

		regexPattern, wildcardNames := patternToRegex(getPattern)

		regex := regexp.MustCompile(regexPattern)
		matches := regex.FindStringSubmatch(uriPath)

		if matches != nil && len(matches) > 1 {
			// first match is the full string, subsequent matches are capture groups
			for i, name := range wildcardNames {
				params[name] = matches[i+1]
			}

			return &uriMatch{
				QueryInputName: queryInputName,
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

// createMessageFromJSON creates a message from the uriMatch given the JSON body in the http request.
func createMessageFromJSON(match *uriMatch, r *http.Request) (gogoproto.Message, error) {
	requestType := gogoproto.MessageType(match.QueryInputName)
	if requestType == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request type")
	}

	msg, ok := reflect.New(requestType.Elem()).Interface().(gogoproto.Message)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to cast to proto message")
	}

	defer r.Body.Close()
	limitedReader := io.LimitReader(r.Body, maxBodySize)
	err := jsonpb.Unmarshal(limitedReader, msg)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return msg, nil

}

// createMessage creates a message from the given uriMatch. If the match has params, the message will be populated
// with the value of those params. Otherwise, an empty message is returned.
func createMessage(match *uriMatch) (gogoproto.Message, error) {
	requestType := gogoproto.MessageType(match.QueryInputName)
	if requestType == nil {
		return nil, status.Error(codes.InvalidArgument, "unknown request type")
	}

	msg, ok := reflect.New(requestType.Elem()).Interface().(gogoproto.Message)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to create message instance")
	}

	// if the uri match has params, we need to populate the message with the values of those params.
	if match.HasParams() {
		// convert flat params map to nested structure
		nestedParams := make(map[string]any)
		for key, value := range match.Params {
			parts := strings.Split(key, ".")
			current := nestedParams

			// step through nested levels
			for i, part := range parts {
				if i == len(parts)-1 {
					// Last part - set the value
					current[part] = value
				} else {
					// continue nestedness
					if _, exists := current[part]; !exists {
						current[part] = make(map[string]any)
					}
					current = current[part].(map[string]any)
				}
			}
		}

		// Configure decoder to handle the nested structure
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:           msg,
			TagName:          "json", // Use json tags as they're simpler
			WeaklyTypedInput: true,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create message instance")
		}

		if err := decoder.Decode(nestedParams); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return msg, nil
}
