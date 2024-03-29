package main

import (
	"html/template"
	"net/http"

	"gitea.com/go-chi/session"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	bolt "go.etcd.io/bbolt"
)

var c Config
var r chi.Router
var entries Entries
var templates *template.Template
var db *bolt.DB

func main() {
	// Load base features.
	c = loadConfig()
	templates = loadTemplates()
	db = loadDB()
	defer db.Close()

	// Load badges, font, etc.
	loadCustom()

	// Get our game jam info if we have a game jam and are missing any pertinent fields.
	loadJamInfo()

	// Get our entries.
	entries = getEntries(c.GameJamID)

	// Set up our HTTP router.
	r = chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(session.Sessioner(session.Options{
		Provider:       "file",
		ProviderConfig: "sessions",
	}))

	// Set up our routes.
	setupRoutes()

	// And listen!
	if err := http.ListenAndServe(c.Address, r); err != nil {
		panic(err)
	}
}
