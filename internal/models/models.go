package models

type Product struct {
	Name        string
	Description string
	Category    string
}

type Service struct {
	Name        string
	Description string
}

type ContactForm struct {
	Name    string
	Email   string
	Message string
}
