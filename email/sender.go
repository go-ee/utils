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
	emailAddress     string
	smtpLogin        string
	smtpPassword     string
	smtpHost         string
	smtpPort         int
	smtpHostWithPort string
}

func NewSender(emailAddress, smtpLogin, smtpPassword, smtpHost string, smtpPort int) *Sender {
	return &Sender{emailAddress, smtpLogin, smtpPassword, smtpHost, smtpPort,
		fmt.Sprintf("%v:%v", smtpHost, smtpPort)}
}

func (o Sender) Send(dest []string, subject, message string) (err error) {
	msg := fmt.Sprintf("From: %v\nTo: %v\nsubject: %v\n%v",
		o.emailAddress, strings.Join(dest, ","), subject, message)

	logrus.Debugf("send, %v, %v", dest, subject)

	if err = smtp.SendMail(o.smtpHostWithPort,
		smtp.PlainAuth("", o.smtpLogin, o.smtpPassword, o.smtpHost),
		o.emailAddress, dest, []byte(msg), &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         o.smtpHost,
		}); err != nil {

		logrus.Warnf("Send, err=%v, %v, %v", err, dest, subject)
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
	ret, err = o.BuildEmail("text/plain", body)
	return
}
