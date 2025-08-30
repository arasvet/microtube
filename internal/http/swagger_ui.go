package http

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed swagger.html
var swaggerHTML embed.FS

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(swaggerHTML, "swagger.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}
