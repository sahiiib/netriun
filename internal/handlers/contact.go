package handlers

import "net/http"

func ContactHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "contact", "Contact Netriun", "Contact Netriun for sales, support and general enquiries.", "https://netriun.com/contact")
}
