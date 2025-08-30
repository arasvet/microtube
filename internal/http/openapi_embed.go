package http

import (
	"embed"
	"net/http"
)

//go:embed openapi.yaml
var openapiFS embed.FS

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	b, err := openapiFS.ReadFile("openapi.yaml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to load openapi"))
		return
	}
	_, _ = w.Write(b)
}
