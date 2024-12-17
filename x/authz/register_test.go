package authz

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	m := GetProtoHTTPGetRuleMapping()
	bz, err := json.Marshal(m)
	require.NoError(t, err)
	os.WriteFile("mapping.json", bz, 0600)
}
