package telemetry

import (
	"context"
	"fmt"
	"net/http"
)

// StandaloneMetrics is a wrapper around application telemetry functionality.
// It should be used independently a server started with telemetry enabled.
// It is not safe to use if a server is started with telemetry enabled.
type StandaloneMetrics struct {
	*Metrics
	mux *http.ServeMux
}

func NewStandaloneMetrics(cfg Config) (*StandaloneMetrics, error) {
	m, err := New(cfg)
	if err != nil {
		return nil, err
	}
	s := &StandaloneMetrics{Metrics: m}

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		//format := strings.TrimSpace(r.FormValue("format"))

		gr, err := s.Gather(FormatPrometheus)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("failed to gather metrics: %s", err)))
			return
		}

		w.Header().Set("Content-Type", gr.ContentType)
		_, _ = w.Write(gr.Metrics)
	})

	return s, nil
}

func (s *StandaloneMetrics) StartServer(ctx context.Context, addr string) {
	go func() {
		server := &http.Server{Addr: addr, Handler: s.mux}
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
		<-ctx.Done()
		if err := server.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
}
