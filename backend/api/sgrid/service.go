package sgrid

import (
	"os"

	"github.com/go-kit/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

/*
	This package controls the email service for sending email verification codes necessary for account verification,
	as well as password resetting
*/

// Interface for the mail service
type Service interface {
	SendMail(mailReq *Mail) error
	NewMail(from string, to string, subject, body string, data *MailData) *Mail
}

// Sendgrid interface
type MailService interface {
	NewMail(from string, to string, subject string, body string, data *MailData) *Mail
	SendMail(mailreq *Mail) error
}

// Structure of the data to be used in the template of the mail
type MailData struct {
	Username string
	Code     string
}

// Email request struct
type Mail struct {
	From    string
	To      string
	Subject string
	Body    string
	Data    *MailData
}

// Sendgrid implementation of mailservice
type SGMailService struct {
	Logger log.Logger
	Client *sendgrid.Client
}

// Returns a new instance of SGMailService
func NewSGMailService(logger log.Logger) *SGMailService {
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	return &SGMailService{
		Logger: logger,
		Client: client,
	}
}

// CreateMail takes in mail request, and constructs a sendgrid mail type
func (ms *SGMailService) SendMail(mailReq *Mail) error {
	from := mail.NewEmail("Greed Finance", mailReq.From)
	subject := mailReq.Subject
	to := mail.NewEmail(mailReq.Data.Username, mailReq.To)
	plainTextContent := mailReq.Body
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, "")

	response, err := ms.Client.Send(message)
	if err != nil {
		ms.Logger.Log(
			"level", "error",
			"msg", "error sending email",
			"err", err,
		)
		return err
	} else {
		ms.Logger.Log(
			"statusCode", response.StatusCode,
			"to", mailReq.To)
	}

	return nil
}

// NewMail returns a new mail request
func (ms *SGMailService) NewMail(from string, to string, subject, body string, data *MailData) *Mail {
	return &Mail{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
		Data:    data,
	}
}
