package context

import (
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"strings"
	"sync"
	"github.com/pkg/errors"
)

type ClientManager struct {
	clients []rpcclient.Client
	currentIndex int
	mutex sync.RWMutex
}

func NewClientManager(nodeURIs string) (*ClientManager,error) {
	if nodeURIs != "" {
		nodeUrlArray := strings.Split(nodeURIs, ",")
		var clients []rpcclient.Client
		for _, url := range nodeUrlArray {
			client := rpcclient.NewHTTP(url, "/websocket")
			clients = append(clients, client)
		}
		mgr := &ClientManager{
			currentIndex: 0,
			clients: clients,
		}
		return mgr, nil
	} else {
		return nil, errors.New("missing node URIs")
	}
}

func (mgr *ClientManager) getClient() rpcclient.Client {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	client := mgr.clients[mgr.currentIndex]
	mgr.currentIndex++
	if mgr.currentIndex >= len(mgr.clients){
		mgr.currentIndex = 0
	}
	return client
}