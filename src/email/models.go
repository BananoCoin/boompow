package email

type ConfirmationEmailData struct {
	ConfirmationCode              string
	ConfirmCodeExpirationDuration int
}
