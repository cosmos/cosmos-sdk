package context

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClientManager(t *testing.T) {
	nodeURIs := "10.10.10.10:26657,20.20.20.20:26657,30.30.30.30:26657"
	clientMgr, err := NewClientManager(nodeURIs)
	assert.Empty(t, err)
	endpoint := clientMgr.getClient()
	assert.NotEqual(t, endpoint, clientMgr.getClient())
	clientMgr.getClient()
	assert.Equal(t, endpoint, clientMgr.getClient())
}