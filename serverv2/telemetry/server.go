package telemetry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/gorilla/mux"
)

type Server struct {
	metrics *telemetry.Metrics
	router  mux.Router
}

func NewServer(r mux.Router, cfg telemetry.Config) (*Server, error) {
	m, err := telemetry.New(cfg)
	if err != nil {
		return nil, err
	}

	metricsHandler := func(w http.ResponseWriter, r *http.Request) {
		format := strings.TrimSpace(r.FormValue("format"))

		gr, err := m.Gather(format)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			bz, err := json.Marshal(errorResponse{Code: 400, Error: fmt.Sprintf("failed to gather metrics: %s", err)})
			if err != nil {
				return
			}
			_, _ = w.Write(bz)

			return
		}

		w.Header().Set("Content-Type", gr.ContentType)
		_, _ = w.Write(gr.Metrics)
	}

	r.HandleFunc("/metrics", metricsHandler).Methods("GET")

	return &Server{
		metrics: m,
		router:  r,
	}, nil
}

// errorResponse defines the attributes of a JSON error response.
type errorResponse struct {
	Code  int    `json:"code,omitempty"`
	Error string `json:"error"`
}
