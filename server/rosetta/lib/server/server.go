package server

import (
	"fmt"
	"net/http"
	"time"

	assert "github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/cosmos/cosmos-sdk/server/rosetta/lib/internal/service"
	crgtypes "github.com/cosmos/cosmos-sdk/server/rosetta/lib/types"
)

const DefaultRetries = 5
const DefaultRetryWait = 5 * time.Second

// Settings define the rosetta server settings
type Settings struct {
	// Network contains the information regarding the network
	Network *types.NetworkIdentifier
	// Client is the online API handler
	Client crgtypes.Client
	// Listen is the address the handler will listen at
	Listen string
	// Offline defines if the rosetta service should be exposed in offline mode
	Offline bool
	// Retries is the number of readiness checks that will be attempted when instantiating the handler
	// valid only for online API
	Retries int
	// RetryWait is the time that will be waited between retries
	RetryWait time.Duration
}

type Server struct {
	h    http.Handler
	addr string
}

func (h Server) Start() error {
	return http.ListenAndServe(h.addr, h.h)
}

func NewServer(settings Settings) (Server, error) {
	asserter, err := assert.NewServer(
		settings.Client.SupportedOperations(),
		true,
		[]*types.NetworkIdentifier{settings.Network},
		nil,
		false,
		"",
	)
	if err != nil {
		return Server{}, fmt.Errorf("cannot build asserter: %w", err)
	}

	var (
		adapter crgtypes.API
	)
	switch settings.Offline {
	case true:
		adapter, err = newOfflineAdapter(settings)
	case false:
		adapter, err = newOnlineAdapter(settings)
	}
	if err != nil {
		return Server{}, err
	}
	h := server.NewRouter(
		server.NewAccountAPIController(adapter, asserter),
		server.NewBlockAPIController(adapter, asserter),
		server.NewNetworkAPIController(adapter, asserter),
		server.NewMempoolAPIController(adapter, asserter),
		server.NewConstructionAPIController(adapter, asserter),
	)

	return Server{
		h:    h,
		addr: settings.Listen,
	}, nil
}

func newOfflineAdapter(settings Settings) (crgtypes.API, error) {
	if settings.Client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	return service.NewOffline(settings.Network, settings.Client)
}

func newOnlineAdapter(settings Settings) (crgtypes.API, error) {
	if settings.Client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if settings.Retries <= 0 {
		settings.Retries = DefaultRetries
	}
	if settings.RetryWait == 0 {
		settings.RetryWait = DefaultRetryWait
	}

	var err error
	err = settings.Client.Bootstrap()
	if err != nil {
		return nil, err
	}

	for i := 0; i < settings.Retries; i++ {
		err = settings.Client.Ready()
		if err != nil {
			time.Sleep(settings.RetryWait)
			continue
		}
		return service.NewOnlineNetwork(settings.Network, settings.Client)
	}
	return nil, fmt.Errorf("maximum number of retries exceeded, last error: %w", err)
}
