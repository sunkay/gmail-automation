package db

type Email struct {
	Id        int64
	Subject   string
	From      string
	To        string
	Cc        string
	Bcc       string
	SentDate  string
	Body      string
	Sender    string
	CreatedAt string
}

type EmailDB interface {
	InsertEmail(email *Email) error
	GetEmails() ([]Email, error)
}
