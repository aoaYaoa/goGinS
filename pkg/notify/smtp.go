package notify

import (
	"context"
	"fmt"
	"net/smtp"
)

// SMTPConfig holds the configuration for the SMTP notifier.
type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

type SMTPNotifier struct {
	host string
	port string
	from string
	auth smtp.Auth
}

// NewSMTP creates an SMTPNotifier using PLAIN auth.
func NewSMTP(cfg SMTPConfig) *SMTPNotifier {
	return &SMTPNotifier{
		host: cfg.Host,
		port: cfg.Port,
		from: cfg.From,
		auth: smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host),
	}
}

func (n *SMTPNotifier) Send(_ context.Context, to, subject, body string) error {
	msg := []byte(
		"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n",
	)
	addr := fmt.Sprintf("%s:%s", n.host, n.port)
	return smtp.SendMail(addr, n.auth, n.from, []string{to}, msg)
}
