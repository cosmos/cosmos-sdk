package swagger

import (
	"embed"
	"net/http"
)

type swaggerHandler struct {
	swaggerFS embed.FS
}

func (h *swaggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'")

	if r.Method == http.MethodOptions {
		return
	}

	http.FileServer(http.FS(h.swaggerFS)).ServeHTTP(w, r)
}
