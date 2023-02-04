package ormtable

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

func (t *tableImpl) QueryMethodHandlers(protoSvcDesc protoreflect.ServiceDescriptor, grpcSvcDesc *grpc.ServiceDesc) error {
	msgName := t.MessageType().Descriptor().Name()
	getMeth := protoSvcDesc.Methods().ByName(protoreflect.Name(fmt.Sprintf("Get%s", msgName)))
	if getMeth == nil {
		return fmt.Errorf("TODO")
	}

	inType, err := t.typeResolver.FindMessageByName(getMeth.Input().FullName())
	if err != nil {
		return errors.Wrapf(err, "TODO")
	}

	var inFieldDescs []protoreflect.FieldDescriptor
	for _, name := range t.primaryKeyIndex.fields.Names() {
		fieldDesc := inType.Descriptor().Fields().ByName(name)
		if fieldDesc == nil {
			return fmt.Errorf("TODO")
		}
		inFieldDescs = append(inFieldDescs, fieldDesc)
	}

	outType, err := t.typeResolver.FindMessageByName(getMeth.Output().FullName())
	if err != nil {
		return errors.Wrapf(err, "TODO")
	}

	outValueField := outType.Descriptor().Fields().ByName("value")
	if outValueField == nil {
		return fmt.Errorf("TODO")
	}

	grpcSvcDesc.Methods = append(grpcSvcDesc.Methods, grpc.MethodDesc{
		MethodName: string(getMeth.Name()),
		Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
			in := inType.New().Interface()
			if err := dec(in); err != nil {
				return nil, err
			}
			info := &grpc.UnaryServerInfo{
				Server:     srv,
				FullMethod: string(getMeth.FullName()),
			}
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				inMsg, ok := req.(proto.Message)
				if !ok {
					return nil, fmt.Errorf("TODO")
				}

				inMsgRef := inMsg.ProtoReflect()
				var inVals []protoreflect.Value
				for _, desc := range inFieldDescs {
					inVals = append(inVals, inMsgRef.Get(desc))
				}

				msg := t.MessageType().New()

				backend, err := t.getBackend(ctx)
				if err != nil {
					return nil, err
				}

				found, err := t.primaryKeyIndex.get(backend, msg.Interface(), inVals)
				if err != nil {
					return nil, err
				}

				if !found {
					return ormerrors.NotFound, nil
				}

				res := outType.New()
				res.Set(outValueField, protoreflect.ValueOfMessage(msg))

				return res.Interface(), nil
			}
			return interceptor(ctx, in, info, handler)
		},
	})

	// TODO unique indexes

	listMeth := protoSvcDesc.Methods().ByName(protoreflect.Name(fmt.Sprintf("List%s", msgName)))
	if listMeth == nil {
		return fmt.Errorf("TODO")
	}

	return nil
}
