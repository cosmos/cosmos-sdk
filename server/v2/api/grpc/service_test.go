package grpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	gogoproto "github.com/cosmos/gogoproto/types/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
)

type MockRequestMessage struct {
	Data string `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *MockRequestMessage) XXX_MessageName() string {
	return "MockRequestMessage"
}
func (m *MockRequestMessage) Reset()         {}
func (m *MockRequestMessage) String() string { return "" }
func (m *MockRequestMessage) ProtoMessage()  {}
func (m *MockRequestMessage) ValidateBasic() error {
	return nil
}

type MockResponseMessage struct {
	Data string `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *MockResponseMessage) Reset()         {}
func (m *MockResponseMessage) String() string { return "" }
func (m *MockResponseMessage) ProtoMessage()  {}
func (m *MockResponseMessage) ValidateBasic() error {
	return nil
}

type mockApp[T transaction.Tx] struct {
	mock.Mock
}

func (m *mockApp[T]) QueryHandlers() map[string]appmodulev2.Handler {
	args := m.Called()
	return args.Get(0).(map[string]appmodulev2.Handler)
}

func (m *mockApp[T]) Query(ctx context.Context, height uint64, msg transaction.Msg) (transaction.Msg, error) {
	args := m.Called(ctx, height, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(transaction.Msg), args.Error(1)
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(app *mockApp[transaction.Tx])
		request       *QueryRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "successful query",
			setupMock: func(app *mockApp[transaction.Tx]) {
				reqMsg := &MockRequestMessage{Data: "request"}
				respMsg := &MockResponseMessage{Data: "response"}

				handlers := map[string]appmodulev2.Handler{
					"/" + proto.MessageName(&MockRequestMessage{}): {
						Func: func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
							return respMsg, nil
						},
						MakeMsg: func() transaction.Msg {
							return reqMsg
						},
						MakeMsgResp: func() transaction.Msg {
							return respMsg
						},
					},
				}
				app.On("QueryHandlers").Return(handlers)
				app.On("Query", mock.Anything, uint64(0), reqMsg).Return(respMsg, nil)
			},

			request:     createTestRequest(t),
			expectError: false,
		},
		{
			name: "handler not found",
			setupMock: func(app *mockApp[transaction.Tx]) {
				handlers := map[string]appmodulev2.Handler{}
				app.On("QueryHandlers").Return(handlers)
			},
			request:       createTestRequest(t),
			expectError:   true,
			expectedError: "rpc error: code = NotFound desc = handler not found for /MockRequestMessage",
		},
		{
			name: "query error",
			setupMock: func(app *mockApp[transaction.Tx]) {
				reqMsg := &MockRequestMessage{Data: "request"}
				respMsg := &MockRequestMessage{Data: "response"}

				handlers := map[string]appmodulev2.Handler{
					"/" + proto.MessageName(&MockRequestMessage{}): {
						Func: func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
							return respMsg, nil
						},
						MakeMsg: func() transaction.Msg {
							return reqMsg
						},
						MakeMsgResp: func() transaction.Msg {
							return respMsg
						},
					},
				}
				app.On("QueryHandlers").Return(handlers)
				app.On("Query", mock.Anything, uint64(0), reqMsg).Return(nil, assert.AnError)
			},
			request:       createTestRequest(t),
			expectError:   true,
			expectedError: fmt.Sprintf("rpc error: code = Internal desc = query failed: %s", assert.AnError.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockApp := &mockApp[transaction.Tx]{}

			if tt.setupMock != nil {
				tt.setupMock(mockApp)
			}

			service := &v2Service{mockApp.QueryHandlers(), mockApp.Query}
			resp, err := service.Query(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
			}

			mockApp.AssertExpectations(t)
		})
	}
}

func TestV2Service_ListQueryHandlers(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(app *mockApp[transaction.Tx])
	}{
		{
			name: "successful list query handlers",
			setupMock: func(app *mockApp[transaction.Tx]) {
				reqMsg := &MockRequestMessage{Data: "request"}
				respMsg := &MockResponseMessage{Data: "response"}

				handlers := map[string]appmodulev2.Handler{
					"/test.Query": {
						Func: func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
							return respMsg, nil
						},
						MakeMsg: func() transaction.Msg {
							return reqMsg
						},
						MakeMsgResp: func() transaction.Msg {
							return respMsg
						},
					},
				}
				app.On("QueryHandlers").Return(handlers)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockApp := &mockApp[transaction.Tx]{}

			if tt.setupMock != nil {
				tt.setupMock(mockApp)
			}

			service := &v2Service{mockApp.QueryHandlers(), mockApp.Query}
			resp, err := service.ListQueryHandlers(context.Background(), &ListQueryHandlersRequest{})

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, resp.Handlers, 1)
			resp.Handlers[0].RequestName = "/MockRequestMessage"
			resp.Handlers[0].ResponseName = "/MockResponseMessage"

			mockApp.AssertExpectations(t)
		})
	}
}

func createTestRequest(t *testing.T) *QueryRequest {
	t.Helper()

	reqMsg := &MockRequestMessage{Data: "request"}
	reqBytes, err := proto.Marshal(reqMsg)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	return &QueryRequest{
		Request: &gogoproto.Any{
			TypeUrl: "/" + proto.MessageName(reqMsg),
			Value:   reqBytes,
		},
	}
}
