package email

import (
	"errors"
	"github.com/go-gomail/gomail"
	"net/mail"
)

type SMTP struct {
	Server   string `yaml:"server", envconfig:"SMTP_SERVER"`
	Port     int    `yaml:"port", envconfig:"SMTP_PORT"`
	User     string `yaml:"user", envconfig:"SMTP_USER"`
	Password string `yaml:"password", envconfig:"SMTP_PASSWORD"`
}

type Sender struct {
	Email    string `yaml:"email", envconfig:"SENDER_EMAIL"`
	Identity string `yaml:"identity", envconfig:"SENDER_IDENTITY"`
	SMTP     SMTP   `yaml:"smtp"`
}

type Message struct {
	To        string
	Subject   string
	HTML      string
	PlainText string
}

func (o *Sender) Send(message *Message) (err error) {

	if err = validate(message); err != nil {
		return
	}

	from := mail.Address{
		Name:    o.Identity,
		Address: o.Email,
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from.String())
	m.SetHeader("To", message.To)
	m.SetHeader("Subject", message.Subject)

	m.SetBody("text/plain", message.PlainText)
	m.AddAlternative("text/html", message.HTML)

	d := gomail.NewDialer(o.SMTP.Server, o.SMTP.Port, o.SMTP.User, o.SMTP.Password)

	err = d.DialAndSend(m)
	return
}

func (o *Sender) Setup() (err error) {
	err = o.validate()
	return
}

func validate(message *Message) (err error) {
	if message.To == "" {
		err = errors.New("no receiver emails configured")
	}
	return
}

func (o *Sender) validate() (err error) {
	if err = o.SMTP.validate(); err != nil {
		return
	}

	if o.Identity == "" {
		err = errors.New("SMTP sender identity is empty")
	} else if o.Email == "" {
		err = errors.New("SMTP sender email is empty")
	}
	return
}

func (o *SMTP) validate() error {
	if o.Server == "" {
		return errors.New("SMTP server config is empty")
	}
	if o.Port == 0 {
		return errors.New("SMTP port config is empty")
	}

	if o.User == "" {
		return errors.New("SMTP user is empty")
	}
	return nil
}
