package email

type SMTPAccessor interface {
	SendEmail(isHTML bool, from string, to []string, subject string, body string) error
}
