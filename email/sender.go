package email

import (
	"errors"
	"github.com/go-gomail/gomail"
	"net/mail"
)

type Sender struct {
	Server         string
	Port           int
	SenderEmail    string
	SenderIdentity string
	SMTPUser       string
	SMTPPassword   string
}

func (o *Sender) Send(to string, subject string, htmlBody string, txtBody  string) error {

	if o.Server == "" {
		return errors.New("SMTP server config is empty")
	}
	if o.Port == 0 {
		return errors.New("SMTP port config is empty")
	}

	if o.SMTPUser == "" {
		return errors.New("SMTP user is empty")
	}

	if o.SenderIdentity == "" {
		return errors.New("SMTP sender identity is empty")
	}

	if o.SenderEmail == "" {
		return errors.New("SMTP sender email is empty")
	}

	if to == "" {
		return errors.New("no receiver emails configured")
	}

	from := mail.Address{
		Name:    o.SenderIdentity,
		Address: o.SenderEmail,
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from.String())
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	m.SetBody("text/plain", txtBody)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(o.Server, o.Port, o.SMTPUser, o.SMTPPassword)

	return d.DialAndSend(m)
}
