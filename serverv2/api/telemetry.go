package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

var ConfigTemplateTelemetry = `
###############################################################################
###                         Telemetry Configuration                         ###
###############################################################################

[telemetry]

# Prefixed with keys to separate services.
service-name = "{{ .Telemetry.ServiceName }}"

# Enabled enables the application telemetry functionality. When enabled,
# an in-memory sink is also enabled by default. Operators may also enabled
# other sinks such as Prometheus.
enabled = {{ .Telemetry.Enabled }}

# Enable prefixing gauge values with hostname.
enable-hostname = {{ .Telemetry.EnableHostname }}

# Enable adding hostname to labels.
enable-hostname-label = {{ .Telemetry.EnableHostnameLabel }}

# Enable adding service to labels.
enable-service-label = {{ .Telemetry.EnableServiceLabel }}

# PrometheusRetentionTime, when positive, enables a Prometheus metrics sink.
prometheus-retention-time = {{ .Telemetry.PrometheusRetentionTime }}

# GlobalLabels defines a global set of name/value label tuples applied to all
# metrics emitted using the wrapper functions defined in telemetry package.
#
# Example:
# [["chain_id", "cosmoshub-1"]]
global-labels = [{{ range $k, $v := .Telemetry.GlobalLabels }}
  ["{{index $v 0 }}", "{{ index $v 1}}"],{{ end }}
]
`

func RegisterMetrics(r mux.Router, cfg telemetry.Config) (*telemetry.Metrics, error) {
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

	return m, nil
}

// errorResponse defines the attributes of a JSON error response.
type errorResponse struct {
	Code  int    `json:"code,omitempty"`
	Error string `json:"error"`
}
