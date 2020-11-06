package email

import (
	"errors"
	"github.com/go-gomail/gomail"
	"net/mail"
)

type Sender struct {
	SenderEmail    string
	SenderIdentity string
	SMTPServer     string
	SMTPPort       int
	SMTPUser       string
	SMTPPassword   string
}

type Message struct {
	To        string
	Subject   string
	HTML      string
	PlainText string
}

func (o *Sender) Send(message *Message) (err error) {

	if err = o.validate(message); err != nil {
		return
	}

	from := mail.Address{
		Name:    o.SenderIdentity,
		Address: o.SenderEmail,
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from.String())
	m.SetHeader("To", message.To)
	m.SetHeader("Subject", message.Subject)

	m.SetBody("text/plain", message.PlainText)
	m.AddAlternative("text/html", message.HTML)

	d := gomail.NewDialer(o.SMTPServer, o.SMTPPort, o.SMTPUser, o.SMTPPassword)

	err = d.DialAndSend(m)
	return
}

func (o *Sender) validate(message *Message) error {
	if o.SMTPServer == "" {
		return errors.New("SMTP server config is empty")
	}
	if o.SMTPPort == 0 {
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

	if message.To == "" {
		return errors.New("no receiver emails configured")
	}
	return nil
}
