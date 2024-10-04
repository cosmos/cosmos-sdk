package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/stf"
)

// MockModule implements both HasMsgHandlers and HasQueryHandlers
type MockModule struct {
	mock.Mock
	appmodulev2.AppModule
}

func (m *MockModule) RegisterMsgHandlers(router appmodulev2.MsgRouter) {
	m.Called(router)
}

func (m *MockModule) RegisterQueryHandlers(router appmodulev2.QueryRouter) {
	m.Called(router)
}

func TestRegisterServices(t *testing.T) {
	mockModule := new(MockModule)

	app := &App[transaction.Tx]{
		msgRouterBuilder:   stf.NewMsgRouterBuilder(),
		queryRouterBuilder: stf.NewMsgRouterBuilder(),
	}

	mm := &MM[transaction.Tx]{
		modules: map[string]appmodulev2.AppModule{
			"mock": mockModule,
		},
	}

	mockModule.On("RegisterMsgHandlers", app.msgRouterBuilder).Once()
	mockModule.On("RegisterQueryHandlers", app.queryRouterBuilder).Once()

	err := mm.RegisterServices(app)

	assert.NoError(t, err)

	mockModule.AssertExpectations(t)
}
