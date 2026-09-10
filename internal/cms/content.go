package cms

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

type Text map[string]string

func (t Text) In(lang string) string {
	if t[lang] != "" {
		return t[lang]
	}
	return t["en"]
}

type Item struct {
	ID          string `json:"id"`
	Name        Text   `json:"name"`
	Description Text   `json:"description"`
	Image       string `json:"image"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	Date        string `json:"date"`
}
type About struct {
	Lead      Text `json:"lead"`
	Ownership Text `json:"ownership"`
	Mission   Text `json:"mission"`
	Approach  Text `json:"approach"`
}
type Content struct {
	Products   []Item            `json:"products"`
	News       []Item            `json:"news"`
	Partners   []Item            `json:"partners"`
	About      About             `json:"about"`
	Recipients map[string]string `json:"recipients"`
}

//go:embed seed.json
var seed []byte

func Default() Content {
	var c Content
	if err := json.Unmarshal(seed, &c); err != nil {
		panic(err)
	}
	return c
}
func ValidURL(raw string, image bool) bool {
	if raw == "" {
		return true
	}
	if image && strings.HasPrefix(raw, "/static/img/") && !strings.ContainsAny(raw, "\\\r\n") && !strings.Contains(raw, "..") {
		return true
	}
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && !strings.ContainsAny(raw, "\r\n")
}
func validText(t Text, required bool, max int) bool {
	if required && strings.TrimSpace(t["en"]) == "" {
		return false
	}
	for k, v := range t {
		if k != "en" && k != "de" && k != "ru" && k != "hy" {
			return false
		}
		if len(v) > max {
			return false
		}
	}
	return true
}
func (c Content) Validate() error {
	for kind, items := range map[string][]Item{"products": c.Products, "news": c.News, "partners": c.Partners} {
		if len(items) > 100 {
			return fmt.Errorf("Too many %s (maximum 100)", kind)
		}
		seen := map[string]bool{}
		for _, item := range items {
			if item.ID == "" || len(item.ID) > 80 || seen[item.ID] {
				return errors.New("Each item needs a unique ID")
			}
			seen[item.ID] = true
			if !validText(item.Name, true, 200) || !validText(item.Description, false, 20000) {
				return errors.New("Provide an English title and keep text within the field limits")
			}
			if !ValidURL(item.URL, false) || !ValidURL(item.Image, true) {
				return errors.New("Links must use HTTPS; images may also use /static/img/ paths")
			}
			if item.Status != "published" && item.Status != "draft" && !(kind == "products" && item.Status == "coming-soon") {
				return errors.New("Invalid publication status")
			}
			if item.Date != "" {
				if _, err := time.Parse("2006-01-02", item.Date); err != nil {
					return errors.New("Invalid publication date")
				}
			}
			if len(item.Image) > 2000 || len(item.URL) > 2000 {
				return errors.New("URL is too long")
			}
			if len(item.Date) > 10 {
				return errors.New("Invalid date")
			}
		}
	}
	for _, t := range []Text{c.About.Lead, c.About.Ownership, c.About.Mission, c.About.Approach} {
		if !validText(t, true, 20000) {
			return errors.New("About fields require English text")
		}
	}
	for _, key := range []string{"sales", "support", "info"} {
		v := c.Recipients[key]
		a, e := mail.ParseAddress(v)
		if e != nil || a.Address != v || strings.ContainsAny(v, "\r\n") {
			return fmt.Errorf("Invalid %s recipient", key)
		}
	}
	return nil
}
