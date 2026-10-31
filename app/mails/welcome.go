package mails

import "github.com/goravel/framework/contracts/mail"

// Welcome is the welcome e-mail sent to a new user: the HTML and text bodies
// are the welcome.tmpl and welcome.txt views under resources/views/mail, both
// rendered with the user's name.
type Welcome struct {
	Name string
	To   string
}

func NewWelcome(name, to string) *Welcome {
	return &Welcome{Name: name, To: to}
}

func (r *Welcome) Attachments() []string { return nil }

func (r *Welcome) Content() *mail.Content {
	return &mail.Content{
		HtmlView: "welcome.tmpl",
		TextView: "welcome.txt",
		With:     map[string]any{"name": r.Name},
	}
}

func (r *Welcome) Envelope() *mail.Envelope {
	return &mail.Envelope{
		To:      []string{r.To},
		Subject: "Welcome to Goravel, " + r.Name,
	}
}

func (r *Welcome) Headers() map[string]string {
	return map[string]string{"X-Mailer": "goravel-example"}
}

func (r *Welcome) Queue() *mail.Queue { return nil }
