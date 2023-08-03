package accounts

import (
	"cosmossdk.io/collections"
	"google.golang.org/protobuf/proto"
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
		ExecuteMsg: ExecuteMessageSchema{},
		QueryMsg:   QueryMessageSchema{},
	}, nil
}

type Schemas struct {
	State      collections.Schema
	InitMsg    InitMsgSchema
	ExecuteMsg ExecuteMessageSchema
	QueryMsg   QueryMessageSchema
}

type ExecuteMessageSchema struct{}
type QueryMessageSchema struct{}

type InitMsgSchema struct {
	DecodeRequest  func([]byte) (proto.Message, error)
	EncodeResponse func(proto.Message) ([]byte, error)
}
