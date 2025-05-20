package utils

import "gopkg.in/gomail.v2"

func SendEmail(to string, subject string, body string) error {
	m := gomail.Message{}
	m.SetHeader("From", "")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, "", "")
	return d.DialAndSend(&m)
}
