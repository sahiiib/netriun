package handlers

import "net/http"

func NewsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "news", "Netriun News", "The latest news and updates from Netriun.", "https://netriun.com/news")
}
