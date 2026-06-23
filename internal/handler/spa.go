package handler

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed all:dist
var dist embed.FS

// SPA serves the Vue single-page application.
type SPA struct {
	handler http.Handler
}

// NewSPA creates a new SPA handler from the embedded dist directory.
func NewSPA() *SPA {
	// Try to load embedded dist. If not available (dev mode), use a fallback.
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Printf("WARNING: embedded dist not found (dev mode?): %v", err)
		return &SPA{handler: nil}
	}
	return &SPA{handler: http.FileServer(http.FS(sub))}
}

// Serve handles SPA routing — static files if they exist, otherwise index.html.
func (s *SPA) Serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.handler != nil {
		s.serveFromEmbed(w, r)
	} else {
		// Dev mode — return a helpful message
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(devFallbackHTML))
	}
}

func (s *SPA) serveFromEmbed(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// If the request is for a static file (has extension) and it exists, serve it
	if strings.Contains(path, ".") {
		s.handler.ServeHTTP(w, r)
		return
	}

	// SPA fallback: serve index.html for all non-API, non-asset routes
	// We serve index.html by rewriting the path
	r.URL.Path = "/"
	s.handler.ServeHTTP(w, r)
}

const devFallbackHTML = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>oathless terminal</title></head>
<body style="background:#191724;color:#e0def4;font-family:monospace;display:flex;align-items:center;justify-content:center;height:100vh;margin:0">
<div style="text-align:center">
<p style="font-size:1.2em;color:#ebbcba">🌸 oathless terminal</p>
<p style="color:#908caa">running in dev mode — frontend not embedded.</p>
<p style="color:#908caa">start the vue dev server: <code style="color:#9ccfd8">cd frontend && npm run dev</code></p>
</div>
</body>
</html>`
