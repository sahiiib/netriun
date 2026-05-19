package handlers

import (
	"html/template"
	"net/http"
)

type PageData struct {
	Title        string
	Description  string
	Canonical    string
	AssetVersion string
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/base.html",
		"web/templates/home.html",
		"web/templates/partials/navbar.html",
		"web/templates/partials/footer.html",
	))

	data := NewPageData(
		"Netriun | Infrastructure Without Friction",
		"Deploy, secure, and scale cloud-native workloads with integrated Kubernetes networking, observability, zero-trust tunnels, and edge connectivity.",
		"https://netriun.com/",
	)

	tmpl.ExecuteTemplate(w, "base", data)
}
