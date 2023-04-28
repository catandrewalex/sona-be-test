package impl

import (
	"fmt"
	"net/smtp"
	"strings"
)

type smtpAccessor struct {
	smtpHost    string
	smtpPort    string
	smtpAddress string

	smtpLogin    string
	smtpPassword string

	defaultFrom string
}

func NewSMTPAccessor(smtpHost string, smtpPort string, smtpLogin string, smtpPassword string, defaultFrom string) *smtpAccessor {
	smtpAddress := fmt.Sprintf("%s:%v", smtpHost, smtpPort)
	return &smtpAccessor{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpAddress:  smtpAddress,
		smtpLogin:    smtpLogin,
		smtpPassword: smtpPassword,
		defaultFrom:  defaultFrom,
	}
}

func (a smtpAccessor) SendEmail(isHTML bool, from string, to []string, subject string, body string) error {
	auth := smtp.PlainAuth("", a.smtpLogin, a.smtpPassword, a.smtpHost)

	if from == "" {
		from = a.defaultFrom
	}

	var mimeType string
	if isHTML {
		mimeType = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
	}

	message := []byte(
		fmt.Sprintf("To: %s\n", strings.Join(to, ", ")) +
			fmt.Sprintf("From: %s\n", from) +
			fmt.Sprintf("Subject: %s\n", subject) +
			mimeType +
			"\n" +
			fmt.Sprintf("%s\n", body),
	)

	err := smtp.SendMail(a.smtpAddress, auth, from, to, message)
	if err != nil {
		return fmt.Errorf("smtp.SendMail(): %w", err)
	}

	return nil
}
