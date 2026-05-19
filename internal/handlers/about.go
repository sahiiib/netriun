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

	data := NewPageData(
		"About Netriun",
		"Learn about Netriun, a cloud-native infrastructure platform for secure Kubernetes networking, edge connectivity, and observability.",
		"https://netriun.com/about",
	)

	tmpl.ExecuteTemplate(w, "base", data)
}
