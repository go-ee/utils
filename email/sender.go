package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/go-ee/utils/net/smtp"
	"github.com/sirupsen/logrus"
	"mime/quotedprintable"
	"strings"
	"time"
)

type Sender struct {
	User             string
	Password         string
	SmtpHost         string
	SmtpPort         int
	smtpHostWithPort string
}

func NewSender(Username, Password string, smtpHost string, smtpPort int) *Sender {
	return &Sender{Username, Password, smtpHost, smtpPort,
		fmt.Sprintf("%v:%v", smtpHost, smtpPort)}
}

func (o Sender) Send(Dest []string, Subject, message string) (err error) {
	msg := fmt.Sprintf("From: %v\nTo: %v\nSubject: %v\n%v",
		o.User, strings.Join(Dest, ","), Subject, message)

	logrus.Debugf("Send, %v, %v", Dest, Subject)

	if err = smtp.SendMail(o.smtpHostWithPort,
		smtp.PlainAuth("", o.User, o.Password, o.SmtpHost),
		o.User, Dest, []byte(msg), &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         o.SmtpHost,
		}); err != nil {

		logrus.Warnf("Send, err=%v, %v, %v", err, Dest, Subject)
	}
	return
}

func (o Sender) BuildEmail(contentType, body string) (ret string, err error) {

	header := make(map[string]string)

	header["Date"] = time.Now().Format(time.RFC1123Z)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("%s; charset=\"utf-8\"", contentType)
	header["Content-Transfer-Encoding"] = "quoted-printable"
	header["Content-Disposition"] = "inline"

	for key, value := range header {
		ret += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	var encodedMessage bytes.Buffer

	finalMessage := quotedprintable.NewWriter(&encodedMessage)
	if _, err = finalMessage.Write([]byte(body)); err == nil {
		return
	}
	if err = finalMessage.Close(); err == nil {
		return
	}
	ret += "\r\n" + encodedMessage.String()
	return
}

func (o *Sender) BuildEmailHTML(body string) (ret string, err error) {
	ret, err = o.BuildEmail("text/html", body)
	return
}

func (o *Sender) BuildEmailPlain(body string) (ret string, err error) {
	ret, err = o.BuildEmail( "text/plain", body)
	return
}
