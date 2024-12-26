package swagger

import (
    "io"
    "net/http"
    "path/filepath"
    "strings"
    "time"

    "github.com/rakyll/statik/fs"
)

// Handler returns an HTTP handler for Swagger UI
func Handler() http.Handler {
    return &swaggerHandler{}
}

type swaggerHandler struct{}

func (h *swaggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Set CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type")

    if r.Method == http.MethodOptions {
        return
    }

    // Get the static file system
    statikFS, err := fs.New()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Process the path
    urlPath := strings.TrimPrefix(r.URL.Path, "/swagger")
    if urlPath == "" || urlPath == "/" {
        urlPath = "/index.html"
    }

    // Open the file
    file, err := statikFS.Open(urlPath)
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
