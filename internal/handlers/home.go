package handlers

import (
	"html/template"
	"net/http"
)

type PageData struct {
	Title string
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/base.html",
		"web/templates/home.html",
		"web/templates/partials/navbar.html",
		"web/templates/partials/footer.html",
	))

	data := PageData{
		Title: "Netriun",
	}

	tmpl.ExecuteTemplate(w, "base", data)
}
