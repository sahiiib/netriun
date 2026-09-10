package main

import (
	"context"
	"log"
	"net/http"
	"netriun.com/internal/cms"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"netriun.com/internal/handlers"
)

func main() {
	var store *cms.Store
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var err error
		store, err = cms.Open(ctx, dsn)
		cancel()
		if err != nil {
			log.Fatal("Unable to initialize content database")
		}
		defer store.DB.Close()
		handlers.ContentStore = store
		go func() {
			ticker := time.NewTicker(time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				store.Cleanup(ctx)
				cancel()
			}
		}()
	} else if os.Getenv("APP_ENV") == "production" {
		log.Fatal("DATABASE_URL is required in production")
	}
	origin := os.Getenv("PUBLIC_ORIGIN")
	if origin == "" {
		origin = "https://netriun.com"
	}
	admin := &cms.Admin{Store: store, Origin: origin, PasswordHash: os.Getenv("ADMIN_PASSWORD_HASH"), Secure: os.Getenv("APP_ENV") == "production"}
	r := mux.NewRouter()
	r.PathPrefix(cms.AdminPath).Handler(admin)
	r.HandleFunc("/api/contact", admin.Contact)
	r.HandleFunc("/news", handlers.NewsHandler).Methods("GET")
	r.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if store != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if store.DB.Ping(ctx) != nil {
				http.Error(w, "Not ready", 503)
				return
			}
		}
		w.Write([]byte("ok"))
	})
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

	err := (&http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}).ListenAndServe()
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
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		if !strings.HasPrefix(r.URL.Path, "/static/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}
