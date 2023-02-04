package ormdb

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/cosmos/cosmos-sdk/orm/internal/codegen"
)

func (m moduleDB) RegisterQueryServices(registrar grpc.ServiceRegistrar) error {
	for _, fileDb := range m.filesById {
		fileDescPath := fileDb.fileDescriptor.Path()
		queryFileDescPath := strings.TrimSuffix(fileDescPath, ".proto")
		queryFileDescPath = queryFileDescPath + "_query.proto"
		resolver := m.options.FileResolver
		if resolver == nil {
			resolver = protoregistry.GlobalFiles
		}

		queryFileDesc, err := resolver.FindFileByPath(queryFileDescPath)
		if err != nil {
			return errors.Wrapf(err, "can't find query proto file for %s, did you forget to run protoc-gen-go-cosmos-orm-proto?", fileDescPath)
		}

		const rerunCosmosOrmProtoGen = "protoc-gen-go-cosmos-orm-proto must be re-run"
		svcName := codegen.QueryServiceName(fileDescPath)
		protoSvcDesc := queryFileDesc.Services().ByName(svcName)
		if protoSvcDesc == nil {
			return fmt.Errorf("can't fine %s, %s seems to be out of date, %s", queryFileDescPath, svcName, rerunCosmosOrmProtoGen)
		}

		grpcSvcDesc := &grpc.ServiceDesc{
			ServiceName: string(protoSvcDesc.FullName()),
		}

		for _, table := range fileDb.tablesByName {
			if handlesQueries, ok := table.(interface {
				QueryMethodHandlers(protoSvcDesc protoreflect.ServiceDescriptor, grpcSvcDesc *grpc.ServiceDesc) error
			}); ok {
				err := handlesQueries.QueryMethodHandlers(protoSvcDesc, grpcSvcDesc)
				if err != nil {
					return err
				}
			}
		}

		registrar.RegisterService(grpcSvcDesc, nil)
	}

	return nil
}
