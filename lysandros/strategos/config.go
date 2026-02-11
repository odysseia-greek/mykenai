package strategos

import (
	"embed"
	"html/template"
	"time"

	"github.com/odysseia-greek/agora/plato/config"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func CreateNewConfig() (*LysandrosHandler, error) {
	resultsDir := config.StringFromEnv("RESULTS_DIR", "./results")

	store := NewStore(resultsDir)
	store.StartPolling(1 * time.Minute)

	tmplIndex := template.Must(template.ParseFS(templatesFS, "templates/index.html"))
	tmplRun := template.Must(template.ParseFS(templatesFS, "templates/run.html"))

	return &LysandrosHandler{
		store:     store,
		tmplIndex: tmplIndex,
		tmplRun:   tmplRun,
	}, nil
}
