package port

type EmailSender interface {
	Send(to, subject, body string) error
}
