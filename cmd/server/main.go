package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"netriun.com/internal/handlers"
)

func main() {
	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./web/static")),
		),
	)

	r.HandleFunc("/", handlers.HomeHandler)
	r.HandleFunc("/about", handlers.AboutHandler)
	r.HandleFunc("/contact", handlers.ContactHandler)
	r.HandleFunc("/products", handlers.ProductsHandler)
	r.HandleFunc("/portal", handlers.PortalHandler)

	log.Println("Netriun running on :8088")

	err := http.ListenAndServe(":8088", r)
	if err != nil {
		panic(err)
	}
}
