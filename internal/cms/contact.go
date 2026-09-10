package cms

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type ContactInput struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
	Website string `json:"website"`
}

func (a *Admin) Contact(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	if r.Header.Get("Origin") != a.Origin {
		fail(w, 403, "Invalid request origin")
		return
	}
	if a.Store == nil {
		fail(w, 503, "Contact form is temporarily unavailable. Please email s.amrei@netriun.com.")
		return
	}
	if !a.Store.Allow(r.Context(), "contact:"+clientKey(r), 5, 10*time.Minute) || !a.Store.Allow(r.Context(), "contact-global", 100, time.Hour) {
		fail(w, 429, "Too many messages. Please try again later.")
		return
	}
	var in ContactInput
	if decode(w, r, &in) != nil {
		fail(w, 400, "Invalid message")
		return
	}
	if in.Website != "" {
		fail(w, 400, "Unable to submit this message")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)
	in.Message = strings.TrimSpace(in.Message)
	email, err := mail.ParseAddress(in.Email)
	if err != nil || email.Address != in.Email || strings.ContainsAny(in.Email+in.Name, "\r\n") || len(in.Email) > 254 || len(in.Name) < 2 || len(in.Name) > 100 || len(in.Message) < 10 || len(in.Message) > 10000 || len(in.Company) > 200 {
		fail(w, 400, "Enter a valid name, email and a message between 10 and 10,000 characters.")
		return
	}
	if in.Topic != "sales" && in.Topic != "support" && in.Topic != "info" {
		fail(w, 400, "Choose a contact topic")
		return
	}
	c, _, err := a.Store.Load(r.Context())
	if err != nil {
		fail(w, 503, "Unable to receive your message. Please try again.")
		return
	}
	recipient := c.Recipients[in.Topic]
	var id int64
	err = a.Store.DB.QueryRow(r.Context(), `INSERT INTO contact_messages(name,email,company,topic,message,recipient) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, in.Name, in.Email, in.Company, in.Topic, in.Message, recipient).Scan(&id)
	if err != nil {
		fail(w, 503, "Unable to receive your message. Please try again.")
		return
	}
	status := "received"
	if MailConfigured() {
		if sendMail(in, recipient) == nil {
			status = "emailed"
		} else {
			status = "email-failed"
		}
		a.Store.DB.Exec(r.Context(), `UPDATE contact_messages SET status=$1 WHERE id=$2`, status, id)
	}
	reply(w, 200, map[string]string{"message": "Your message has been received. Thank you for contacting Netriun."})
}
func sendMail(in ContactInput, to string) error {
	host, port, from := os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"), os.Getenv("SMTP_FROM")
	if port == "" {
		port = "587"
	}
	sender, err := mail.ParseAddress(from)
	if err != nil || sender.Address != from || strings.ContainsAny(from, "\r\n") {
		return fmt.Errorf("invalid sender")
	}
	addr := net.JoinHostPort(host, port)
	dialer := net.Dialer{Timeout: 10 * time.Second}
	config := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	var conn net.Conn
	if port == "465" {
		conn, err = tls.DialWithDialer(&dialer, "tcp", addr, config)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if port != "465" {
		if err = client.StartTLS(config); err != nil {
			return err
		}
	}
	if user := os.Getenv("SMTP_USERNAME"); user != "" {
		if err = client.Auth(smtp.PlainAuth("", user, os.Getenv("SMTP_PASSWORD"), host)); err != nil {
			return err
		}
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "From: %s\r\nTo: %s\r\nReply-To: %s\r\nSubject: Netriun contact: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nName: %s\r\nCompany: %s\r\n\r\n%s\r\n", from, to, in.Email, in.Topic, in.Name, strings.ReplaceAll(strings.ReplaceAll(in.Company, "\r", " "), "\n", " "), strings.ReplaceAll(strings.ReplaceAll(in.Message, "\r\n", "\n"), "\n", "\r\n"))
	if err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

// MailConfigured avoids attempting delivery when only part of the SMTP settings exists.
func MailConfigured() bool {
	for _, key := range []string{"SMTP_HOST", "SMTP_FROM", "SMTP_USERNAME", "SMTP_PASSWORD"} {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return false
		}
	}
	return true
}
