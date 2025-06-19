package email

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/gomail.v2"
)

type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	From         string
	To           []string
	Subject      string
	Body         string
}

type Sender struct {
	cfg Config
}

func NewSender(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Send(data []byte, filename string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.From)
	m.SetHeader("To", s.cfg.To...)
	m.SetHeader("Subject", s.cfg.Subject)
	m.SetBody("text/plain", s.cfg.Body)

	m.Attach(filename, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := io.Copy(w, bytes.NewReader(data))
		return err
	}))

	d := gomail.NewDialer(
		s.cfg.SMTPHost,
		s.cfg.SMTPPort,
		s.cfg.SMTPUsername,
		s.cfg.SMTPPassword,
	)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}