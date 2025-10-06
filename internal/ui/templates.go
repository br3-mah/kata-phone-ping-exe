package ui

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*.html templates/components/*.html
var templateFS embed.FS

var tmpl = template.Must(template.New("root").ParseFS(
	templateFS,
	"templates/*.html",
	"templates/components/*.html",
))

func RenderIndex(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "index.html", nil)
}
