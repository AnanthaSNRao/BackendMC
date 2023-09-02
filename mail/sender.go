package mail

import "net/smtp"

const (
	smptAuthAddress  = "smtp.gmail.com"
	smptSeverAddress = "smtp.gmail.com:587"
)

type EmailSender interface {
	SendEmail(subject string, body string, to []string) error
}

type GmailSender struct {
	name          string
	email         string
	emailPassword string
}

func NewGmailSender(fromName string, fromAddr string, emailPassword string) EmailSender {
	return &GmailSender{
		name:          fromName,
		email:         fromAddr,
		emailPassword: emailPassword,
	}
}

func (sender *GmailSender) SendEmail(subject string, body string, to []string) error {

	auth := smtp.PlainAuth("", sender.email, sender.emailPassword, smptAuthAddress)
	msg := []byte(subject + "\r\n" + body)
	err := smtp.SendMail(smptSeverAddress, auth, sender.email, to, msg)
	if err != nil {
		return err
	}
	return nil
}
