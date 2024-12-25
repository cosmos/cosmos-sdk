package swagger

import (
    "net/http"
    "path"
    "path/filepath"
    "strings"

    "github.com/rakyll/statik/fs"
)

// Handler returns an HTTP handler that serves the Swagger UI files
func Handler(statikFS http.FileSystem) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // If the path is empty or "/", show index.html
        if r.URL.Path == "/" || r.URL.Path == "" {
            r.URL.Path = "/index.html"
        }

        // Clearing the path
        urlPath := path.Clean(r.URL.Path)

        // Opening the file from statikFS
        f, err := statikFS.Open(urlPath)
        if err != nil {
            w.WriteHeader(http.StatusNotFound)
            return
        }
        defer f.Close()

        // Determining the content-type
        ext := strings.ToLower(filepath.Ext(urlPath))
        switch ext {
        case ".html":
            w.Header().Set("Content-Type", "text/html")
        case ".css":
            w.Header().Set("Content-Type", "text/css")
        case ".js":
            w.Header().Set("Content-Type", "application/javascript")
        case ".json":
            w.Header().Set("Content-Type", "application/json")
        }

        http.ServeContent(w, r, urlPath, time.Time{}, f)
    }
} 
