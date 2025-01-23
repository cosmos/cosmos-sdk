package swagger

import (
	"io/fs"
	"net/http"
)

type swaggerHandler struct {
	swaggerFS fs.FS
}

func (h *swaggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'")

	if r.Method == http.MethodOptions {
		return
	}

	root, err := fs.Sub(h.swaggerFS, "swagger-ui")
	if err != nil {
		http.Error(w, "failed to get swagger-ui from fs", http.StatusInternalServerError)
	}

	staticServer := http.FileServer(http.FS(root))
	http.StripPrefix("/swagger/", staticServer).ServeHTTP(w, r)
}
