package cms

import (
	"bytes"
	"context"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestContentValidation(t *testing.T) {
	c := Default()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"javascript:alert(1)", "http://example.com", "//example.com", "https://user:pass@example.com", "/static/img/../key"} {
		if ValidURL(raw, true) {
			t.Errorf("accepted unsafe URL %q", raw)
		}
	}
	c.Products[0].URL = "javascript:alert(1)"
	if c.Validate() == nil {
		t.Fatal("unsafe content accepted")
	}
	c = Default()
	c.Recipients["sales"] = "a@example.com\r\nBcc: b@example.com"
	if c.Validate() == nil {
		t.Fatal("header injection accepted")
	}
	if (Text{"en": "Fallback"}).In("hy") != "Fallback" {
		t.Fatal("missing language fallback")
	}
}
func TestEditorIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	s, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer s.DB.Close()
	if _, err = s.DB.Exec(ctx, `TRUNCATE admin_sessions,request_limits,contact_messages; UPDATE site_content SET revision=1`); err != nil {
		t.Fatal(err)
	}
	seed, _ := json.Marshal(Default())
	s.DB.Exec(ctx, `UPDATE site_content SET body=$1`, seed)
	hash, _ := bcrypt.GenerateFromPassword([]byte("a-long-test-password"), bcrypt.MinCost)
	a := &Admin{Store: s, Origin: "https://netriun.com", PasswordHash: string(hash), Secure: true}
	request := func(path, method string, body any, cookie *http.Cookie, csrf, origin string) *httptest.ResponseRecorder {
		var raw []byte
		if body != nil {
			raw, _ = json.Marshal(body)
		}
		r := httptest.NewRequest(method, AdminPath+path, bytes.NewReader(raw))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		r.Header.Set("X-CSRF-Token", csrf)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		a.ServeHTTP(w, r)
		return w
	}
	if w := request("/api/content", "GET", nil, nil, "", ""); w.Code != 401 {
		t.Fatal("unauthorized read", w.Code)
	}
	login := map[string]string{"username": "admin", "password": "a-long-test-password"}
	if w := request("/api/login", "POST", login, nil, "", "https://evil.example"); w.Code != 403 {
		t.Fatal("cross-origin login", w.Code)
	}
	w := request("/api/login", "POST", login, nil, "", a.Origin)
	if w.Code != 200 {
		t.Fatal("login", w.Code, w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("insecure session cookie")
	}
	var session map[string]string
	json.Unmarshal(w.Body.Bytes(), &session)
	csrf := session["csrf"]
	// A second server shares sessions and sees all committed edits.
	b := &Admin{Store: s, Origin: a.Origin, PasswordHash: string(hash), Secure: true}
	r := httptest.NewRequest("GET", AdminPath+"/api/session", nil)
	r.AddCookie(cookie)
	w = httptest.NewRecorder()
	b.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal("session not shared")
	}
	c, rev, err := s.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	c.About.Lead["en"] = "Updated from editor"
	payload := map[string]any{"content": c, "revision": rev}
	if w = request("/api/content", "PUT", payload, cookie, "bad", a.Origin); w.Code != 403 {
		t.Fatal("CSRF accepted")
	}
	if w = request("/api/content", "PUT", payload, cookie, csrf, a.Origin); w.Code != 200 {
		t.Fatal("save", w.Code, w.Body.String())
	}
	if w = request("/api/content", "PUT", payload, cookie, csrf, a.Origin); w.Code != 409 {
		t.Fatal("stale save accepted")
	}
	reopened, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	saved, _, err := reopened.Load(ctx)
	reopened.DB.Close()
	if err != nil || saved.About.Lead["en"] != "Updated from editor" {
		t.Fatal("content not persistent")
	}
	c.Products[0].URL = "javascript:alert(1)"
	if w = request("/api/content", "PUT", map[string]any{"content": c, "revision": rev + 1}, cookie, csrf, a.Origin); w.Code != 400 {
		t.Fatal("unsafe link saved")
	}
	t.Setenv("SMTP_HOST", "")
	input := ContactInput{Name: "Example User", Email: "test@example.com", Topic: "sales", Message: "Please tell me about Nexus."}
	raw, _ := json.Marshal(input)
	r = httptest.NewRequest("POST", "/api/contact", bytes.NewReader(raw))
	r.Header.Set("Origin", a.Origin)
	r.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	a.Contact(w, r)
	if w.Code != 200 {
		t.Fatal("contact", w.Code, w.Body.String())
	}
	w = request("/api/messages", "GET", nil, cookie, "", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "test@example.com") {
		t.Fatal("message not stored")
	}
	if s.Allow(ctx, "limit-test", 1, time.Minute) != true || s.Allow(ctx, "limit-test", 1, time.Minute) != false {
		t.Fatal("rate limit not enforced")
	}
	if w = request("/api/logout", "POST", map[string]string{}, cookie, csrf, a.Origin); w.Code != 200 {
		t.Fatal("logout", w.Code)
	}
	if w = request("/api/content", "GET", nil, cookie, "", ""); w.Code != 401 {
		t.Fatal("revoked session works")
	}
}
