package cms

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const AdminPath = "/neditport2065"
const cookieName = "netriun_editor"

type Admin struct {
	Store                *Store
	Origin, PasswordHash string
	Secure               bool
}

func randomToken() string {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func digest(s string) string { v := sha256.Sum256([]byte(s)); return hex.EncodeToString(v[:]) }
func reply(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, code int, s string) { reply(w, code, map[string]string{"error": s}) }
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return errors.New("Expected JSON")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	if d.Decode(new(any)) != io.EOF {
		return errors.New("Invalid request")
	}
	return nil
}
func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return digest(host)
}
func (a *Admin) session(r *http.Request) (string, string, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil || len(cookie.Value) != 64 {
		return "", "", false
	}
	key := digest(cookie.Value)
	var csrf string
	err = a.Store.DB.QueryRow(r.Context(), `SELECT csrf FROM admin_sessions WHERE token=$1 AND expires>now()`, key).Scan(&csrf)
	return key, csrf, err == nil
}
func (a *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' https:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
	if a.Store == nil || a.PasswordHash == "" {
		http.Error(w, "Editor is not configured", 503)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, AdminPath)
	if path == "" || path == "/" {
		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}
		http.ServeFile(w, r, "web/admin/index.html")
		return
	}
	if path == "/admin.js" || path == "/admin.css" {
		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}
		http.ServeFile(w, r, "web/admin"+path)
		return
	}
	if r.Method != "GET" {
		if r.Header.Get("Origin") != a.Origin {
			fail(w, 403, "Invalid request origin")
			return
		}
	}
	if path == "/api/login" && r.Method == "POST" {
		if !a.Store.Allow(r.Context(), "login:"+clientKey(r), 10, 15*time.Minute) || !a.Store.Allow(r.Context(), "login-global", 100, 15*time.Minute) {
			fail(w, 429, "Too many attempts. Please try again later.")
			return
		}
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if decode(w, r, &input) != nil || len(input.Password) > 72 {
			fail(w, 400, "Invalid login request")
			return
		}
		err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(input.Password))
		if err != nil || input.Username != "admin" {
			fail(w, 401, "Invalid username or password")
			return
		}
		token, csrf := randomToken(), randomToken()
		expires := time.Now().Add(8 * time.Hour)
		_, err = a.Store.DB.Exec(r.Context(), `INSERT INTO admin_sessions(token,csrf,expires) VALUES($1,$2,$3)`, digest(token), csrf, expires)
		if err != nil {
			fail(w, 503, "Unable to start session")
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: token, Path: AdminPath, HttpOnly: true, Secure: a.Secure, SameSite: http.SameSiteStrictMode, Expires: expires, MaxAge: 28800})
		reply(w, 200, map[string]string{"csrf": csrf})
		return
	}
	key, csrf, ok := a.session(r)
	if !ok {
		fail(w, 401, "Please sign in")
		return
	}
	if r.Method != "GET" && subtle.ConstantTimeCompare([]byte(r.Header.Get("X-CSRF-Token")), []byte(csrf)) != 1 {
		fail(w, 403, "Invalid session token")
		return
	}
	switch {
	case path == "/api/session" && r.Method == "GET":
		reply(w, 200, map[string]any{"csrf": csrf, "emailConfigured": MailConfigured()})
	case path == "/api/logout" && r.Method == "POST":
		if _, err := a.Store.DB.Exec(r.Context(), `DELETE FROM admin_sessions WHERE token=$1`, key); err != nil {
			fail(w, 503, "Unable to sign out")
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Path: AdminPath, MaxAge: -1, HttpOnly: true, Secure: a.Secure, SameSite: http.SameSiteStrictMode})
		reply(w, 200, map[string]bool{"ok": true})
	case path == "/api/content" && r.Method == "GET":
		c, rev, err := a.Store.Load(r.Context())
		if err != nil {
			fail(w, 503, "Content unavailable")
			return
		}
		reply(w, 200, map[string]any{"content": c, "revision": rev})
	case path == "/api/content" && r.Method == "PUT":
		var input struct {
			Content  Content `json:"content"`
			Revision int64   `json:"revision"`
		}
		if err := decode(w, r, &input); err != nil {
			fail(w, 400, "Invalid or oversized content")
			return
		}
		if err := input.Content.Validate(); err != nil {
			fail(w, 400, err.Error())
			return
		}
		if err := a.Store.Save(r.Context(), input.Content, input.Revision); err != nil {
			if errors.Is(err, ErrConflict) {
				fail(w, 409, err.Error())
			} else {
				fail(w, 503, "Save failed. Your changes have not been published.")
			}
			return
		}
		reply(w, 200, map[string]any{"revision": input.Revision + 1})
	case path == "/api/messages" && r.Method == "GET":
		rows, err := a.Store.DB.Query(r.Context(), `SELECT id,created,name,email,company,topic,message,recipient,status FROM contact_messages ORDER BY id DESC LIMIT 100`)
		if err != nil {
			fail(w, 503, "Messages unavailable")
			return
		}
		defer rows.Close()
		messages := []map[string]any{}
		for rows.Next() {
			var id int64
			var created time.Time
			var name, email, company, topic, message, recipient, status string
			if rows.Scan(&id, &created, &name, &email, &company, &topic, &message, &recipient, &status) != nil {
				fail(w, 503, "Messages unavailable")
				return
			}
			messages = append(messages, map[string]any{"id": id, "created": created, "name": name, "email": email, "company": company, "topic": topic, "message": message, "recipient": recipient, "status": status})
		}
		if rows.Err() != nil {
			fail(w, 503, "Messages unavailable")
			return
		}
		reply(w, 200, messages)
	default:
		http.NotFound(w, r)
	}
}
