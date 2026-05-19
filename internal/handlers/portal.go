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

	data := NewPageData(
		"Netriun Customer Portal",
		"Access the Netriun customer portal for infrastructure status, support, account activity, and cloud-native operations.",
		"https://netriun.com/portal",
	)

	tmpl.ExecuteTemplate(w, "base", data)
}
