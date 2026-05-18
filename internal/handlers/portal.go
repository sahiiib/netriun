package handlers

import (
	"html/template"
	"net/http"
)

func PortalHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/base.html",
		"web/templates/portal.html",
		"web/templates/partials/navbar.html",
		"web/templates/partials/footer.html",
	))

	data := PageData{
		Title: "Customer Portal",
	}

	tmpl.ExecuteTemplate(w, "base", data)
}
