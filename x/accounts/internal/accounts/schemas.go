package accounts

import (
	"cosmossdk.io/collections"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

func NewSchemas(
	stateSchemaBuilder *collections.SchemaBuilder,
	initRouter *InitRouter,
	executeRouter *ExecuteRouter,
	queryRouter *QueryRouter,
) (*Schemas, error) {
	stateSchema, err := stateSchemaBuilder.Build()
	if err != nil {
		return nil, err
	}
	return &Schemas{
		State:      stateSchema,
		InitMsg:    *initRouter.schema,
		ExecuteMsg: *executeRouter.schema,
	}, nil
}

type Schemas struct {
	State      collections.Schema
	InitMsg    InitMsgSchema
	ExecuteMsg ExecuteMessageSchema
	QueryMsg   QueryMessageSchema
}

type ExecuteMessageSchema struct {
	DecodeRequest  func([]byte) (proto.Message, error)
	EncodeResponse func(proto.Message) ([]byte, error)

	requestDecoders map[protoreflect.FullName]func(anyPB *anypb.Any) (proto.Message, error)
}
type QueryMessageSchema struct{}

type InitMsgSchema struct {
	DecodeRequest  func([]byte) (proto.Message, error)
	EncodeResponse func(proto.Message) ([]byte, error)
}
