package handlers

import "net/http"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "home", "Netriun | Infrastructure Without Friction", "Deploy, secure, and scale cloud-native workloads with integrated Kubernetes networking, observability, and edge connectivity.", "https://netriun.com/")
}
