package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)
	expPool := initialPool()

	//check that the empty keeper loads the default
	resPool := keeper.getPool(ctx)
	assert.Equal(t, expPool, resPool)

	//modify a params, save, and retrieve
	expPool.TotalSupply = 777
	keeper.setPool(ctx, expPool)
	resPool = keeper.getPool(ctx)
	assert.Equal(t, expPool, resPool)
}
