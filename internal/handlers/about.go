package handlers

import "net/http"

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "about", "About Netriun", "Learn about Netriun, a brand owned by Bamshi LLC in Yerevan, Armenia.", "https://netriun.com/about")
}
