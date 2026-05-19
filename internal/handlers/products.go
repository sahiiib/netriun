package handlers

import (
	"html/template"
	"net/http"
)

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/base.html",
		"web/templates/products.html",
		"web/templates/partials/navbar.html",
		"web/templates/partials/footer.html",
	))

	data := NewPageData(
		"Netriun Products",
		"Explore Netriun products for Kubernetes networking, zero-trust service exposure, infrastructure automation, and multi-cloud operations.",
		"https://netriun.com/products",
	)

	tmpl.ExecuteTemplate(w, "base", data)
}
