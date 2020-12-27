package email

import (
	"testing"
)

func TestSendBySender(t *testing.T) {
	err := sendBySender()
	if err != nil {
		t.Error(err)
	}
}

func sendBySender() (err error) {
	var sender Sender
	if err = LoadSenderConfig("sender-config.yml", &sender); err != nil {
		return
	}

	if err = sender.Setup(); err != nil {
		return
	}

	err = sender.Send(messageDefault())

	if err != nil {
		return
	}
	return
}

func messageDefault() *Message {
	return &Message{
		To:        "eoeisler@gmail.com",
		Subject:   "TestSender",
		HTML:      "</br></br>TestSender",
		PlainText: "TestSender",
	}
}
