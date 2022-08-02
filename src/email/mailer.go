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

	"github.com/bbedward/boompow-server-ng/src/config"
	"github.com/bbedward/boompow-server-ng/src/utils"
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
		Name:    "Banano",
		Address: "noreply@plausible.banano.cc",
	}
	to := mail.Address{
		Name:    "Your Name",
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
func SendConfirmationEmail(destination string, token string) error {
	// Load template
	t, err := loadEmailTemplate("confirmemail.html")
	if err != nil {
		return err
	}

	// Encode URL params
	// !  TODO - real URL eventually
	urlParam := url.QueryEscape(fmt.Sprintf(`query verifyEmail{
		verifyEmail(input:{email:"%s", token:"%s"})
	}`, destination, token))

	// Populate template
	templateData := ConfirmationEmailData{
		ConfirmationCode:              fmt.Sprintf("http://localhost:8080/graphql?query=%s", urlParam),
		ConfirmCodeExpirationDuration: config.EMAIL_CONFIRMATION_TOKEN_VALID_MINUTES,
	}

	return sendEmail(
		destination,
		"Confirm your email address for your BoomPOW Account",
		t, templateData,
	)
}
