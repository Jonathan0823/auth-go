package utils

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

func SendEmail(to string, subject string, body string) error {
	email := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")
	if email == "" || password == "" {
		return fmt.Errorf("email or password not set in environment variables")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	fmt.Printf("Sending email to %s with subject %s\n", to, subject)

	d := gomail.NewDialer("smtp.gmail.com", 587, email, password)
	return d.DialAndSend(m)
}
