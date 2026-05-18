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

	data := PageData{
		Title: "Products",
	}

	tmpl.ExecuteTemplate(w, "base", data)
}
