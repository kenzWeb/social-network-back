package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

type EmailSender interface {
	Send(to, subject, body string) error
}

type SMTPSender struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func (s *SMTPSender) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	msg := []byte("From: " + s.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" + body + "\r\n")

	if s.Port == 465 {
	}
	return smtp.SendMail(addr, auth, s.From, []string{to}, msg)
}

func tlsConfig(host string) *tls.Config {
	return &tls.Config{ServerName: host}
}
