package email

type ConfirmationEmailData struct {
	ConfirmationLink              string
	ConfirmCodeExpirationDuration int
	IsProvider                    bool
}
