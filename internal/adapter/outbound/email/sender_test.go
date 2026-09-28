package email

import (
	"testing"
)

func TestSenderRequiresCredentials(t *testing.T) {
	if err := NewSender("", "").Send("user@example.com", "subject", "body"); err == nil {
		t.Fatal("Send succeeded without credentials")
	}
}
