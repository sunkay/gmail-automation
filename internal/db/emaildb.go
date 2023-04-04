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
	Labels    string
	Read      bool
	Deleted   bool
	CreatedAt string
}

type EmailDB interface {
	InsertEmail(email *Email) (id int64, err error)
	GetEmails(tableName string) ([]Email, error)
	UpdateEmailReadStatus(id int64, read bool) error
	UpdateEmailLabels(id int64, labels string) error
	GetEmail(tableName string, subject string, from string, to string, sentDate string) (Email, error)

	// batch update methods
	InsertEmails(emails []Email) (int64, error)
	//UpdateEmailReadStatuses(ids []int64, read bool) error
	//UpdateEmailLabelses(ids []int64, labels string) error

	InsertDeletedEmails(emails []Email) (int64, error)
	//GetDeletedEmails() ([]Email, error)
	//GetDeletedEmail(subject string, from string, to string, sentDate string) (Email, error)
}
