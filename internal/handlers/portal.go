package handlers

import "net/http"

func PortalHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "portal", "Netriun Customer Portal", "Access the Netriun customer portal.", "https://netriun.com/portal")
}
