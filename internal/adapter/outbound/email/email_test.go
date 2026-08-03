package email

import (
	"testing"
)

func TestSenderRequiresCredentials(t *testing.T) {
	t.Setenv("EMAIL", "")
	t.Setenv("PASSWORD", "")
	if err := NewSender().Send("user@example.com", "subject", "body"); err == nil {
		t.Fatal("Send succeeded without credentials")
	}
}
