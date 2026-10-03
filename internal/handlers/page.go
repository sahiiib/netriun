package handlers

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net/http"
	"netriun.com/internal/cms"
	"time"
)

const assetVersion = "20261003-about-theme-1"

var ContentStore *cms.Store

type PageData struct {
	Title, Description, Canonical, AssetVersion, Lang string
	Content                                           cms.Content
}

func NewPageData(title, description, canonical string) PageData {
	return PageData{Title: title, Description: description, Canonical: canonical, AssetVersion: assetVersion, Lang: "en", Content: cms.Default()}
}
func renderPage(w http.ResponseWriter, r *http.Request, page, title, description, canonical string) {
	data := NewPageData(title, description, canonical)
	if c, err := r.Cookie("netriun-language"); err == nil {
		switch c.Value {
		case "en", "de", "ru", "hy":
			data.Lang = c.Value
		}
	}
	if ContentStore != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		c, _, err := ContentStore.Load(ctx)
		if err != nil {
			http.Error(w, "Content temporarily unavailable", 503)
			return
		}
		data.Content = c
	}
	funcs := template.FuncMap{"local": func(t cms.Text) string { return t.In(data.Lang) }}
	tmpl, err := template.New("page").Funcs(funcs).ParseFiles("web/templates/base.html", "web/templates/"+page+".html", "web/templates/partials/navbar.html", "web/templates/partials/footer.html", "web/templates/partials/products.html")
	if err != nil {
		log.Print("Template parse failed: ", err)
		http.Error(w, "Page unavailable", 500)
		return
	}
	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		log.Print("Template render failed: ", err)
		http.Error(w, "Page unavailable", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
