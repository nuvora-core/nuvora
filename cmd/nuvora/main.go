package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nuvora-core/nuvora/internal/config"
	"github.com/nuvora-core/nuvora/internal/handlers/providers"
	"github.com/nuvora-core/nuvora/internal/store"
	"github.com/nuvora-core/nuvora/web"
)

func main() {
	// flags
	port := flag.String("port", ":9090", "port for run server")

	// Init project
	err := config.Init()

	if err != nil {
		fmt.Println("config: init project error")
		fmt.Println("err: ", err)
		os.Exit(1)
	}

	// nuvora path
	nuvoraPath := config.GetNuvoraPath()

	// create(connect) database
	err = store.ConnectDB(
		filepath.Join(nuvoraPath, "nuvora.db"),
	)

	if err != nil {
		fmt.Println("store: connect/create database error")
		fmt.Println("err: ", err)
		os.Exit(1)
	}

	defer store.CloseDB()

	// create tables
	err = store.CreateTables()

	if err != nil {
		fmt.Println("store: create tables error")
		fmt.Println("err: ", err)
		os.Exit(1)
	}

	// embedded frontend
	webHandler, err := web.Handler()

	if err != nil {
		fmt.Println("web: load embedded frontend error")
		fmt.Println("err: ", err)
		os.Exit(1)
	}

	// router
	router := chi.NewRouter()

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/provider", providers.ConnectProvider)

		r.Get("/provider", providers.GetYourProviders)

		r.Delete("/provider/{id}", providers.DeleteProvider)
	})

	// everything else is served by the embedded frontend
	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)

			return
		}

		webHandler.ServeHTTP(w, r)
	}))

	http.ListenAndServe(*port, router)
}
