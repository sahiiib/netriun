package handlers

import (
	"html/template"
	"net/http"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/base.html",
		"web/templates/about.html",
		"web/templates/partials/navbar.html",
		"web/templates/partials/footer.html",
	))

	data := PageData{
		Title: "About Netriun",
	}

	tmpl.ExecuteTemplate(w, "base", data)
}
