package handlers

import "net/http"

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "products", "Netriun Products", "Explore the Netriun platform and upcoming products.", "https://netriun.com/products")
}
