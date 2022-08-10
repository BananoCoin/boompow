package email

type ConfirmationEmailData struct {
	ConfirmationLink              string
	ConfirmCodeExpirationDuration int
	IsProvider                    bool
}

type ResetPasswordEmailData struct {
	ResetPasswordLink string
}

type ConfirmServiceEmailData struct {
	EmailAddress       string
	ServiceName        string
	ServiceWebsite     string
	ApproveServiceLink string
}
