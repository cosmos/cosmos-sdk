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

    // Add security headers
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "DENY")
    w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'")

    if r.Method == http.MethodOptions {
        return
    }

    // Process and validate the path
    urlPath := strings.TrimPrefix(r.URL.Path, "/swagger")
    if urlPath == "" || urlPath == "/" {
        urlPath = "/index.html"
    }

    // Clean the path before validation
    urlPath = filepath.Clean(urlPath)

    // Validate path before any operations
    if strings.Contains(urlPath, "..") || strings.Contains(urlPath, "//") || strings.Contains(urlPath, "\\") {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
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
