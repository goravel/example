package mails

import "github.com/goravel/framework/contracts/mail"

// Invoice is queued instead of sent inline: Queue() asks for the "mails" queue
// on the "database" connection, and the HTML body is written inline (no view).
type Invoice struct {
	Number string
	To     string
}

func NewInvoice(number, to string) *Invoice {
	return &Invoice{Number: number, To: to}
}

func (r *Invoice) Attachments() []string { return nil }

func (r *Invoice) Content() *mail.Content {
	return &mail.Content{Html: "<p>Invoice " + r.Number + " is ready.</p>"}
}

func (r *Invoice) Envelope() *mail.Envelope {
	return &mail.Envelope{
		To:      []string{r.To},
		Subject: "Invoice " + r.Number,
	}
}

func (r *Invoice) Headers() map[string]string { return nil }

func (r *Invoice) Queue() *mail.Queue {
	return &mail.Queue{Connection: "database", Queue: "mails"}
}
