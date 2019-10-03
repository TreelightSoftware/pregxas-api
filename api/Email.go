package api

import (
	"fmt"

	"gopkg.in/mailgun/mailgun-go.v1"
)

// SendEmail sends an email. The body is the complete HTML body, so the expectation is that the
// header, translation, and footer are all present and combined. This file will typically not have a ton of test coverage as we don't want to
// actually send an email through MailGun
func SendEmail(to string, subject string, body string) (string, string, error) {
	//we don't want to throw an error, we just want to return right out if we are in a no-send
	if !Config.MailShouldSend {
		return "", "", nil
	}

	from := Config.MailFromAddress
	domain := Config.MailgunDomain
	privateKey := Config.MailgunPrivateKey
	publicKey := Config.MailgunPublicKey

	if domain == "" || privateKey == "" || publicKey == "" {
		panic("Mailgun not set up!")
	}
	mg := mailgun.NewMailgun(domain, privateKey, publicKey)
	message := mailgun.NewMessage(
		from,
		subject,
		body,
		to)
	message.SetHtml(body)
	resp, id, err := mg.Send(message)
	return resp, id, err
}

// SendEmailToGroup sends an email to a group of email addresses. This file will typically not have a ton of test coverage as we don't want to
// actually send an email through MailGun
func SendEmailToGroup(to []string, subject string, body string, asBCC bool) (string, string, error) {
	//we don't want to throw an error, we just want to return right out if we are in a no-send
	if !Config.MailShouldSend {
		return "", "", nil
	}

	from := Config.MailFromAddress
	domain := Config.MailgunDomain
	privateKey := Config.MailgunPrivateKey
	publicKey := Config.MailgunPublicKey

	if domain == "" || privateKey == "" || publicKey == "" {
		panic("Mailgun not set up!")
	}
	mg := mailgun.NewMailgun(domain, privateKey, publicKey)
	message := mailgun.NewMessage(
		from,
		subject,
		body,
		Config.MailFromAddress)
	message.SetHtml(body)
	for i := range to {
		if asBCC {
			message.AddBCC(to[i])
		} else {
			message.AddRecipient(to[i])
		}
	}
	resp, id, err := mg.Send(message)
	return resp, id, err
}

// GenerateEmail generates an email by combining the inserted body with a header and footer setup. The
// community ID is primarily for the logo if it is present, which is a forthcoming feature
func GenerateEmail(communityID int64, body string) string {
	// TODO: Build out the header and footer for the org

	header := ""
	footer := fmt.Sprintf(`<p>The Pregxas Team</p><p>If you believe you received this message in error, please forward this message to %s.</p>`, Config.MailFromAddress)
	return fmt.Sprintf("%s%s%s", header, body, footer)
}
