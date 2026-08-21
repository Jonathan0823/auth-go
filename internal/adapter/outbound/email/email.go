package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type sender struct {
	address  string
	password string
}

func NewSender(address, password string) *sender {
	return &sender{address: address, password: password}
}

func (s *sender) Send(to, subject, body string) error {
	if s.address == "" || s.password == "" {
		return fmt.Errorf("email credentials not set")
	}
	m := gomail.NewMessage()
	m.SetHeader("From", s.address)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, s.address, s.password)
	return d.DialAndSend(m)
}
