package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/mail"
	"net/smtp"
	"net/url"
	"path/filepath"
	"runtime"

	"github.com/bananocoin/boompow/apps/server/src/config"
	"github.com/bananocoin/boompow/apps/server/src/models"
	"github.com/bananocoin/boompow/libs/utils"
	"github.com/golang/glog"
)

// Returns path of an email template by name
func getTemplatePath(name string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Join(filepath.Dir(b), "templates", name)
	return basepath
}

// Send an email with given parameters
func sendEmail(destination string, subject string, t *template.Template, templateData interface{}) error {
	// Get credentials
	smtpCredentials := utils.GetSmtpConnInformation()
	if smtpCredentials == nil {
		errMsg := "SMTP Credentials misconfigured, not sending email"
		glog.Errorf(errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	// Send email
	auth := smtp.PlainAuth("", smtpCredentials.Username, smtpCredentials.Password, smtpCredentials.Server)
	from := mail.Address{
		Name:    "BoomPoW (Banano)",
		Address: "noreply@mail.banano.cc",
	}
	to := mail.Address{
		Address: destination,
	}

	title := subject

	var body bytes.Buffer

	if err := t.ExecuteTemplate(&body, "base", templateData); err != nil {
		glog.Errorf("Error creating email template  %s", err)
		return err
	}

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = title
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString(body.Bytes())

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", smtpCredentials.Server, smtpCredentials.Port),
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
	)
	if err != nil {
		glog.Errorf("Error sending email  %s", err)
		return err
	}

	return nil
}

// Load email template from file
func loadEmailTemplate(templateName string) (*template.Template, error) {
	// Load template
	t, err := template.New("").ParseFiles(getTemplatePath(templateName), getTemplatePath("base.html"))
	if err != nil {
		glog.Errorf("Failed to load email template %s, %s", templateName, err)
		return nil, err
	}
	return t, nil
}

// Send email with link to verify user's email address
func SendConfirmationEmail(destination string, userType models.UserType, token string) error {
	// Load template
	t, err := loadEmailTemplate("confirmemail.html")
	if err != nil {
		return err
	}

	// Encode URL params
	urlParam := url.QueryEscape(fmt.Sprintf(`query verifyEmail{
		verifyEmail(input:{email:"%s", token:"%s"})
	}`, destination, token))

	// Populate template
	templateData := ConfirmationEmailData{
		ConfirmationLink:              fmt.Sprintf("https://boompow.banano.cc/graphql?query=%s", urlParam),
		ConfirmCodeExpirationDuration: config.EMAIL_CONFIRMATION_TOKEN_VALID_MINUTES,
		IsProvider:                    userType == models.PROVIDER,
	}

	return sendEmail(
		destination,
		"Confirm your email address for your BoomPOW Account",
		t, templateData,
	)
}

// Send email with link to reset user's password
func SendResetPasswordEmail(destination string, token string) error {
	// Load template
	t, err := loadEmailTemplate("resetpassword.html")
	if err != nil {
		return err
	}

	// Populate template
	// ! TODO - this is something we'll want to to link to the frontend - which will have a form for a new password
	templateData := ResetPasswordEmailData{
		ResetPasswordLink: "https://boompow.banano.cc/notimplemented",
	}

	return sendEmail(
		destination,
		"Reset the password for your BoomPOW Account",
		t, templateData,
	)
}

// Send email with link to authorize service
func SendAuthorizeServiceEmail(email string, name string, website string, token string) error {
	// Load template
	t, err := loadEmailTemplate("confirmservice.html")
	if err != nil {
		return err
	}

	// Encode URL params
	urlParam := url.QueryEscape(fmt.Sprintf(`query verifyService{
		verifyService(input:{email:"%s", token:"%s"})
	}`, email, token))

	// Populate template
	templateData := ConfirmServiceEmailData{
		ServiceName:        name,
		EmailAddress:       email,
		ServiceWebsite:     website,
		ApproveServiceLink: fmt.Sprintf("https://boompow.banano.cc/graphql?query=%s", urlParam),
	}

	return sendEmail(
		"hello@appditto.com",
		"A service has requested access to BoomPoW",
		t, templateData,
	)
}

// Send email letting service know they are approved
func SendServiceApprovedEmail(email string) error {
	// Load template
	t, err := loadEmailTemplate("serviceapproved.html")
	if err != nil {
		return err
	}

	// Populate template
	templateData := map[string]string{}
	return sendEmail(
		email,
		"You have been authorized to use BoomPoW!",
		t, templateData,
	)
}
