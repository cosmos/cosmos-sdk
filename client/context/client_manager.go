package context

import (
	"github.com/pkg/errors"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"strings"
	"sync"
)

// ClientManager is a manager of a set of rpc clients to full nodes.
// This manager can do load balancing upon these rpc clients.
type ClientManager struct {
	clients      []rpcclient.Client
	currentIndex int
	mutex        sync.Mutex
}

// NewClientManager create a new ClientManager
func NewClientManager(nodeURIs string) (*ClientManager, error) {
	if nodeURIs != "" {
		nodeURLArray := strings.Split(nodeURIs, ",")
		var clients []rpcclient.Client
		for _, url := range nodeURLArray {
			client := rpcclient.NewHTTP(url, "/websocket")
			clients = append(clients, client)
		}
		mgr := &ClientManager{
			currentIndex: 0,
			clients:      clients,
		}
		return mgr, nil
	}
	return nil, errors.New("missing node URIs")
}

func (mgr *ClientManager) getClient() rpcclient.Client {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	client := mgr.clients[mgr.currentIndex]
	mgr.currentIndex++
	if mgr.currentIndex >= len(mgr.clients) {
		mgr.currentIndex = 0
	}
	return client
}
