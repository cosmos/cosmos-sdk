package grpcgateway

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
)

const MaxBodySize = 1 << 20 // 1 MB

var (
	_ http.Handler = &protoHandler[transaction.Tx]{}

	wildcardRegex = regexp.MustCompile(`\{([^}]*)\}`)
)

// queryMetadata holds information related to handling gateway queries.
type queryMetadata struct {
	// queryInputProtoName is the proto name of the query's input type.
	msg gogoproto.Message
	// wildcardKeyNames are the wildcard key names from the query's HTTP annotation.
	// for example /foo/bar/{baz}/{qux} would produce []string{"baz", "qux"}
	// this is used for building the query's path parameter map.
	wildcardKeyNames []string
}

// mountHTTPRoutes registers handlers for from proto HTTP annotations to the http.ServeMux, using runtime.ServeMux as a fallback/
// last ditch effort router.
func mountHTTPRoutes[T transaction.Tx](httpMux *http.ServeMux, fallbackRouter *runtime.ServeMux, am appmanager.AppManager[T]) error {
	annotationMapping, err := newHTTPAnnotationMapping()
	if err != nil {
		return err
	}
	annotationToMetadata, err := annotationsToQueryMetadata(annotationMapping)
	if err != nil {
		return err
	}
	registerMethods[T](httpMux, am, fallbackRouter, annotationToMetadata)
	return nil
}

// registerMethods registers the endpoints specified in the annotation mapping to the http.ServeMux.
func registerMethods[T transaction.Tx](mux *http.ServeMux, am appmanager.AppManager[T], fallbackRouter *runtime.ServeMux, annotationToMetadata map[string]queryMetadata) {
	// register the fallback handler. this will run if the mux isn't able to get a match from the registrations below.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fallbackRouter.ServeHTTP(w, r)
	})

	uris := slices.Sorted(maps.Keys(annotationToMetadata))
	for _, uri := range uris {
		queryMD := annotationToMetadata[uri]
		mux.Handle(uri, &protoHandler[T]{
			msg:              queryMD.msg,
			fallbackRouter:   fallbackRouter,
			appManager:       am,
			wildcardKeyNames: queryMD.wildcardKeyNames,
		})
	}
}

// protoHandler handles turning data in http.Request to the gogoproto.Message
type protoHandler[T transaction.Tx] struct {
	// msg is the gogoproto message type.
	msg gogoproto.Message
	// wildcardKeyNames are the wildcard key names, if any, specified in the http annotation. (i.e. /foo/bar/{baz})
	wildcardKeyNames []string
	// fallbackRouter is the canonical gRPC gateway runtime.ServeMux, used as a fallback if the query does not have a handler in AppManager.
	fallbackRouter *runtime.ServeMux
	// appManager is used to route queries.
	appManager appmanager.AppManager[T]
}

func (p *protoHandler[T]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	in, out := runtime.MarshalerForRequest(p.fallbackRouter, request)

	// we clone here as handlers are concurrent and using p.msg would trample.
	msg := gogoproto.Clone(p.msg)

	// extract path parameters.
	params := make(map[string]string)
	for _, wildcardKeyName := range p.wildcardKeyNames {
		params[wildcardKeyName] = request.PathValue(wildcardKeyName)
	}

	inputMsg, err := p.populateMessage(request, in, msg, params)
	if err != nil {
		// the errors returned from the message creation return status errors. no need to make one here.
		runtime.HTTPError(request.Context(), p.fallbackRouter, out, writer, request, err)
		return
	}

	// get the height from the header.
	var height uint64
	heightStr := request.Header.Get(GRPCBlockHeightHeader)
	heightStr = strings.Trim(heightStr, `\"`)
	if heightStr != "" && heightStr != "latest" {
		height, err = strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
			runtime.HTTPError(request.Context(), p.fallbackRouter, out, writer, request, status.Errorf(codes.InvalidArgument, "invalid height in header: %s", heightStr))
			return
		}
	}

	responseMsg, err := p.appManager.Query(request.Context(), height, inputMsg)
	if err != nil {
		// if we couldn't find a handler for this request, just fall back to the fallbackRouter.
		if strings.Contains(err.Error(), "no handler") {
			p.fallbackRouter.ServeHTTP(writer, request)
		} else {
			// for all other errors, we just return the error.
			runtime.HTTPError(request.Context(), p.fallbackRouter, out, writer, request, err)
		}
		return
	}

	runtime.ForwardResponseMessage(request.Context(), p.fallbackRouter, out, writer, request, responseMsg)
}

func (p *protoHandler[T]) populateMessage(req *http.Request, marshaler runtime.Marshaler, input gogoproto.Message, pathParams map[string]string) (gogoproto.Message, error) {
	// see if we have path params to populate the message with.
	if len(pathParams) > 0 {
		for pathKey, pathValue := range pathParams {
			if err := runtime.PopulateFieldFromPath(input, pathKey, pathValue); err != nil {
				return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to populate field %s with value %s: %w", pathKey, pathValue, err).Error())
			}
		}
	}

	// handle query parameters.
	if err := req.ParseForm(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	filter := filterFromPathParams(pathParams)
	err := runtime.PopulateQueryParameters(input, req.Form, filter)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	// see if we have a body to unmarshal.
	if req.ContentLength > 0 {
		if req.ContentLength > MaxBodySize {
			return nil, status.Errorf(codes.InvalidArgument, "request body too large: %d bytes, max=%d", req.ContentLength, MaxBodySize)
		}

		// this block of code ensures that the body can be re-read. this is needed as if the query fails in the
		// app's query handler, we need to pass the request back to the fallbackRouter, which needs to be able to
		// read the body again.
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		if err = marshaler.NewDecoder(bytes.NewReader(bodyBytes)).Decode(input); err != nil && !errors.Is(err, io.EOF) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
	}

	return input, nil
}

func filterFromPathParams(pathParams map[string]string) *utilities.DoubleArray {
	var prefixPaths [][]string

	for k := range pathParams {
		prefixPaths = append(prefixPaths, []string{k})
	}

	return utilities.NewDoubleArray(prefixPaths)
}

// newHTTPAnnotationMapping returns a mapping of RPC Method HTTP GET annotation to the RPC Handler's Request Input type full name.
//
// example: "/cosmos/auth/v1beta1/account_info/{address}":"cosmos.auth.v1beta1.Query.AccountInfo"
func newHTTPAnnotationMapping() (map[string]string, error) {
	protoFiles, err := gogoproto.MergedRegistry()
	if err != nil {
		return nil, err
	}

	annotationToQueryInputName := make(map[string]string)
	protoFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Services().Len(); i++ {
			serviceDesc := fd.Services().Get(i)
			for j := 0; j < serviceDesc.Methods().Len(); j++ {
				methodDesc := serviceDesc.Methods().Get(j)
				httpExtension := proto.GetExtension(methodDesc.Options(), annotations.E_Http)
				if httpExtension == nil {
					continue
				}

				httpRule, ok := httpExtension.(*annotations.HttpRule)
				if !ok || httpRule == nil {
					continue
				}
				queryInputName := string(methodDesc.Input().FullName())
				httpRules := append(httpRule.GetAdditionalBindings(), httpRule)
				for _, rule := range httpRules {
					if httpAnnotation := rule.GetGet(); httpAnnotation != "" {
						annotationToQueryInputName[fixCatchAll(httpAnnotation)] = queryInputName
					}
					if httpAnnotation := rule.GetPost(); httpAnnotation != "" {
						annotationToQueryInputName[fixCatchAll(httpAnnotation)] = queryInputName
					}
					if httpAnnotation := rule.GetPut(); httpAnnotation != "" {
						annotationToQueryInputName[fixCatchAll(httpAnnotation)] = queryInputName
					}
					if httpAnnotation := rule.GetPatch(); httpAnnotation != "" {
						annotationToQueryInputName[fixCatchAll(httpAnnotation)] = queryInputName
					}
					if httpAnnotation := rule.GetDelete(); httpAnnotation != "" {
						annotationToQueryInputName[fixCatchAll(httpAnnotation)] = queryInputName
					}
				}
			}
		}
		return true
	})
	return annotationToQueryInputName, nil
}

var catchAllRegex = regexp.MustCompile(`\{([^=]+)=\*\*\}`)

// fixCatchAll replaces grpc gateway catch all syntax with net/http syntax.
//
// {foo=**} -> {foo...}
func fixCatchAll(uri string) string {
	return catchAllRegex.ReplaceAllString(uri, `{$1...}`)
}

// annotationsToQueryMetadata takes annotations and creates a mapping of URIs to queryMetadata.
func annotationsToQueryMetadata(annotations map[string]string) (map[string]queryMetadata, error) {
	annotationToMetadata := make(map[string]queryMetadata)
	for uri, queryInputName := range annotations {
		// extract the proto message type.
		msgType := gogoproto.MessageType(queryInputName)
		if msgType == nil {
			continue
		}
		msg, ok := reflect.New(msgType.Elem()).Interface().(gogoproto.Message)
		if !ok {
			return nil, fmt.Errorf("query input type %q does not implement gogoproto.Message", queryInputName)
		}
		annotationToMetadata[uri] = queryMetadata{
			msg:              msg,
			wildcardKeyNames: extractWildcardKeyNames(uri),
		}
	}
	return annotationToMetadata, nil
}

// extractWildcardKeyNames extracts the wildcard key names from the uri annotation.
//
// example:
// "/hello/{world}" -> []string{"world"}
// "/hello/{world}/and/{friends} -> []string{"world", "friends"}
// "/hello/world" -> []string{}
func extractWildcardKeyNames(uri string) []string {
	matches := wildcardRegex.FindAllStringSubmatch(uri, -1)
	var extracted []string
	for _, match := range matches {
		// match[0] is the full string including braces (i.e. "{bar}")
		// match[1] is the captured group (i.e. "bar")
		// we also need to handle the catch-all case with URI's like "bar..." and
		// transform them to just "bar".
		extracted = append(extracted, strings.TrimRight(match[1], "."))
	}
	return extracted
}
