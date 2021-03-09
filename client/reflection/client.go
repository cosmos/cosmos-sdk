package reflection

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"

	tmrpc "github.com/tendermint/tendermint/rpc/client"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	"google.golang.org/grpc"
	grpcreflectionv1alpha "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/client/reflection/unstructured"
)

type Client struct {
	sdkReflect  reflection.ReflectionServiceClient
	grpcReflect grpcreflectionv1alpha.ServerReflectionClient

	tmClient tmrpc.Client

	msgs     map[string]ServiceMsg
	queriers methodsMap
}

func (c *Client) ListQueries() []Queries {
	var q []Queries
	for name, method := range c.queriers {
		q = append(q, Queries{
			Service:  method.ServiceName,
			Method:   name,
			Request:  (string)(method.Request.FullName()),
			Response: (string)(method.Response.FullName()),
		})
	}

	return q
}

func (c *Client) Query(ctx context.Context, method string, request unstructured.Object) (resp proto.Message, err error) {
	desc, exists := c.queriers[method]
	if !exists {
		return nil, fmt.Errorf("unknown method: %s", method)
	}

	reqProto, err := request.MarshalFromDescriptor(desc.Request)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request to proto message: %w", err)
	}

	b, err := proto.Marshal(reqProto)

	tmResp, err := c.tmClient.ABCIQuery(ctx, method, b)
	if err != nil {
		return nil, err
	}

	resp = dynamicpb.NewMessage(desc.Response)
	return resp, proto.Unmarshal(tmResp.Response.Value, resp)

}

func NewClient(grpcEndpoint, tmEndpoint string) (*Client, error) {
	conn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	sdkReflect := reflection.NewReflectionServiceClient(conn)
	grpcReflect := grpcreflectionv1alpha.NewServerReflectionClient(conn)

	tmRpc, err := tmhttp.New(tmEndpoint, "")
	if err != nil {
		return nil, err
	}

	c := &Client{
		sdkReflect:  sdkReflect,
		grpcReflect: grpcReflect,
		tmClient:    tmRpc,
		msgs:        nil,
		queriers:    make(methodsMap),
	}
	err = c.init()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) init() error {
	queries, err := c.sdkReflect.ListQueryServices(context.TODO(), nil)
	if err != nil {
		return err
	}

	svcPerFile := make(map[string][]string)

	for _, q := range queries.Queries {
		_, exists := svcPerFile[q.ProtoFile]
		if !exists {
			svcPerFile[q.ProtoFile] = nil
		}

		svcPerFile[q.ProtoFile] = append(svcPerFile[q.ProtoFile], q.ServiceName)
	}

	svcDescriptors := make([][]byte, 0, len(svcPerFile))

	for file := range svcPerFile {
		rawDesc, err := c.sdkReflect.ResolveService(context.TODO(), &reflection.ResolveServiceRequest{FileName: file})
		if err != nil {
			return err
		}

		svcDescriptors = append(svcDescriptors, rawDesc.RawDescriptor)
	}

	err = c.buildQueries(svcDescriptors)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) buildQueries(rawDescs [][]byte) error {
	for _, rawDesc := range rawDescs {
		fileDesc, err := decodeDescriptor(rawDesc)
		if err != nil {
			return err
		}
		queries, err := queryFromFileDesc(fileDesc)
		if err != nil {
			return err
		}

		err = c.queriers.merge(queries)
		if err != nil {
			return err
		}
	}

	return nil
}

func decodeDescriptor(rawDesc []byte) (protoreflect.FileDescriptor, error) {
	// build is used multiple times and needs a recover because the protoimpl builder
	// expects the bytes provided to be valid proto files, which cannot always assume
	// to be true.
	build := func(rawDesc []byte) (desc protoreflect.FileDescriptor, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("unable to build descriptor: %#v", r)
			}
		}()

		builder := protoimpl.DescBuilder{
			RawDescriptor: rawDesc,
		}
		return builder.Build().File, nil
	}
	buf := bytes.NewBuffer(rawDesc)
	unzip, err := gzip.NewReader(buf)
	if err != nil {
		return build(rawDesc)
	}

	unzippedDesc, err := ioutil.ReadAll(unzip)
	if err != nil {
		return nil, err
	}

	return build(unzippedDesc)
}

func queryFromFileDesc(file protoreflect.FileDescriptor) (methodsMap, error) {
	svcs := file.Services()

	m := make(methodsMap)

	for i := 0; i < svcs.Len(); i++ {
		svcDesc := svcs.Get(i)
		methodsFromSvc, err := methodsFromSvcDesc(svcDesc)
		if err != nil {
			return nil, err
		}

		err = m.merge(methodsFromSvc)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func methodsFromSvcDesc(desc protoreflect.ServiceDescriptor) (methodsMap, error) {
	m := make(methodsMap)

	methods := desc.Methods()
	for i := 0; i < methods.Len(); i++ {
		method := methods.Get(i)
		methodName := fmt.Sprintf("/%s/%s", desc.FullName(), method.Name()) // TODO: why in the sdk we broke standard grpc query method invocation naming?
		err := m.insert(methodName, method)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
