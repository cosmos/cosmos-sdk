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

	if r.Method == http.MethodOptions {
		return
	}

	http.StripPrefix("/swagger/", http.FileServer(http.FS(h.swaggerFS))).ServeHTTP(w, r)
}
