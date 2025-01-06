package swagger

import (
    "io"
    "net/http"
    "path/filepath"
    "strings"
    "time"
)

type swaggerHandler struct {
    swaggerFS http.FileSystem
}

func (h *swaggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Set minimal CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET")

    if r.Method == http.MethodOptions {
        return
    }

    // Process the path
    urlPath := strings.TrimPrefix(r.URL.Path, "/swagger")
    if urlPath == "" || urlPath == "/" {
        urlPath = "/index.html"
    }

    // Open the file
    file, err := h.swaggerFS.Open(urlPath)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    defer file.Close()

    // Set the content-type
    ext := filepath.Ext(urlPath)
    if ct := getContentType(ext); ct != "" {
        w.Header().Set("Content-Type", ct)
    }

    // Set caching headers
    w.Header().Set("Cache-Control", "public, max-age=31536000")
    w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

    // Serve the file
    http.ServeContent(w, r, urlPath, time.Now(), file.(io.ReadSeeker))
}

// getContentType returns the content-type for a file extension
func getContentType(ext string) string {
    switch strings.ToLower(ext) {
    case ".html":
        return "text/html"
    case ".css":
        return "text/css"
    case ".js":
        return "application/javascript"
    case ".json":
        return "application/json"
    case ".png":
        return "image/png"
    case ".jpg", ".jpeg":
        return "image/jpeg"
    case ".svg":
        return "image/svg+xml"
    default:
        return ""
    }
}
