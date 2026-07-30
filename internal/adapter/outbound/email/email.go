package email

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

type sender struct{}

func NewSender() *sender {
	return &sender{}
}

func (s *sender) Send(to, subject, body string) error {
	email := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")
	if email == "" || password == "" {
		return fmt.Errorf("email credentials not set")
	}
	m := gomail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, email, password)
	return d.DialAndSend(m)
}
