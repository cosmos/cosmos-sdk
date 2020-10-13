package baseapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgService(t *testing.T) {
	qr := NewMsgServiceRouter()
	interfaceRegistry := testdata.NewTestInterfaceRegistry()
	qr.SetInterfaceRegistry(interfaceRegistry)
	testdata.RegisterMsgServer(qr, testdata.MsgImpl{})
	helper := &MsgServiceTestHelper{
		MsgServiceRouter: qr,
		ctx:              sdk.Context{}.WithContext(context.Background()),
	}
	client := testdata.NewMsgClient(helper)

	res, err := client.CreateDog(context.Background(), &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "spot"}})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, "spot", res.Name)

	require.Panics(t, func() {
		_, _ = client.CreateDog(context.Background(), nil)
	})
}
