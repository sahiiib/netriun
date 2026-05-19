package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	"netriun.com/internal/handlers"
)

func main() {
	r := mux.NewRouter()
	r.Use(noCacheDynamic)

	staticFiles := http.FileServer(http.Dir("./web/static"))
	r.PathPrefix("/static/").Handler(cacheStatic(http.StripPrefix("/static/", staticFiles)))

	r.HandleFunc("/", handlers.HomeHandler)
	r.HandleFunc("/about", handlers.AboutHandler)
	r.HandleFunc("/contact", handlers.ContactHandler)
	r.HandleFunc("/products", handlers.ProductsHandler)
	r.HandleFunc("/portal", handlers.PortalHandler)
	r.HandleFunc("/healthz", handlers.HealthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	addr := ":" + port
	log.Printf("Netriun running on %s", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}

func cacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "v=") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}

func noCacheDynamic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/static/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}
