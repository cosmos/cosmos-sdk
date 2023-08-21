package implementation

import "google.golang.org/protobuf/types/known/wrapperspb"

var _ Account = (*TestAccount)(nil)

type TestAccount struct{}

func (TestAccount) RegisterInitHandler(builder *InitBuilder) {
	RegisterInitHandler(builder, func(req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		return &wrapperspb.StringValue{Value: req.Value + "init-echo"}, nil
	})
}

func (TestAccount) RegisterExecuteHandlers(builder *ExecuteBuilder) {
	RegisterExecuteHandler(builder, func(req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		return &wrapperspb.StringValue{Value: req.Value + "execute-echo"}, nil
	})

	RegisterExecuteHandler(builder, func(req *wrapperspb.BytesValue) (*wrapperspb.BytesValue, error) {
		return &wrapperspb.BytesValue{Value: append(req.Value, "bytes-execute-echo"...)}, nil
	})
}

func (TestAccount) RegisterQueryHandlers(builder *QueryBuilder) {
	RegisterQueryHandler(builder, func(req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		return &wrapperspb.StringValue{Value: req.Value + "query-echo"}, nil
	})
	RegisterQueryHandler(builder, func(req *wrapperspb.BytesValue) (*wrapperspb.BytesValue, error) {
		return &wrapperspb.BytesValue{Value: append(req.Value, "bytes-query-echo"...)}, nil
	})
}
