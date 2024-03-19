package web

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templates *template.Template

func Mount(mux *http.ServeMux) error {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return err
	}

	mux.Handle("/{$}", page(index))

	return nil
}

func page(handler func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			slog.Error("Error from handler", "error", err)
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
		}
	})
}

func index(w http.ResponseWriter, r *http.Request) error {
	if err := templates.ExecuteTemplate(w, "index.html", map[string]any{"motd": "You have very many money"}); err != nil {
		return fmt.Errorf("Could not render template: %w", err)
	}
	return nil
}
