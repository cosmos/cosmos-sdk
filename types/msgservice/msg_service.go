package msgservice

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"

	"cosmossdk.io/core/registry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterMsgServiceDesc registers all type_urls from Msg services described
// in `sd` into the registry.
func RegisterMsgServiceDesc(registrar registry.InterfaceRegistrar, sd *grpc.ServiceDesc) {
	for _, method := range sd.Methods {
		_, _ = method.Handler(nil, context.Background(), func(req any) error {
			msg, ok := req.(proto.Message)
			if !ok {
				panic(fmt.Errorf("expected proto.Message, got %T", req))
			}
			registrar.RegisterImplementations((*sdk.Msg)(nil), msg)
			return nil
		}, nil)
	}
}

func unzip(b []byte) []byte {
	if b == nil {
		return nil
	}
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	unzipped, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return unzipped
}
