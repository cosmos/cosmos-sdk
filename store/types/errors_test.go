package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func subst(str string) string {
	return strings.Replace(str, "sdk", "query", -1)
}

func errEqual(t *testing.T, err1 types.Error, err2 sdk.Error, ty string) {
	errMsg := "%s mismatch in %s between store/types and types"

	require.EqualValues(t, err1.Code(), err2.Code(), errMsg, "Code", ty)

	require.EqualValues(t, err1.ABCILog(), subst(err2.ABCILog()), errMsg, "ABCILog", ty)
	require.EqualValues(t, err1.QueryResult().Code, err2.QueryResult().Code, errMsg, "QueryResult.Code", ty)
	require.EqualValues(t, err1.QueryResult().Log, subst(err2.QueryResult().Log), errMsg, "QueryResult.Log", ty)
	require.EqualValues(t, err1.Error(), subst(err2.Error()), errMsg, "Error", ty)
	require.EqualValues(t, types.CodeToDefaultMsg(err1.Code()), sdk.CodeToDefaultMsg(err2.Code()), errMsg, "CodeToDefaultMsg", ty)
}

func TestErrMatch(t *testing.T) {
	msg := "test error msg"

	errEqual(t, types.ErrInternal(msg), sdk.ErrInternal(msg), "ErrInternal")
	errEqual(t, types.ErrTxDecode(msg), sdk.ErrTxDecode(msg), "ErrTxDecode")
	errEqual(t, types.ErrUnknownRequest(msg), sdk.ErrUnknownRequest(msg), "ErrUnknownRequest")
}
