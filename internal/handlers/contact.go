package handlers

import (
	"html/template"
	"net/http"
)

func ContactHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/base.html",
		"web/templates/contact.html",
		"web/templates/partials/navbar.html",
		"web/templates/partials/footer.html",
	))

	data := PageData{
		Title: "Contact Netriun",
	}

	tmpl.ExecuteTemplate(w, "base", data)
}
