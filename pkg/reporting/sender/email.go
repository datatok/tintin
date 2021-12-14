package sender

import (
	"crypto/tls"
	"strings"
	"time"

	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
)

type EmailSender struct {
	Server string
}

type Email struct {
	To, From, Title, Body string
}

func (sender *EmailSender) Send(e Email) {
	ee := &email.Email{
		ReplyTo:     nil,
		From:        e.From,
		To:          strings.Split(e.To, ","),
		Bcc:         nil,
		Cc:          nil,
		Subject:     e.Title,
		Text:        []byte(""),
		HTML:        []byte(e.Body),
		Sender:      "",
		Headers:     nil,
		Attachments: nil,
		ReadReceipt: nil,
	}

	logrus.Infof("Sending email to %s", e.To)

	p, _ := email.NewPool(
		sender.Server,
		1,
		nil,
		&tls.Config{
			InsecureSkipVerify: true,
		},
	)

	err := p.Send(ee, 10*time.Second)

	if err != nil {
		logrus.Errorf("Error send email: %s", err.Error())
	}
}
