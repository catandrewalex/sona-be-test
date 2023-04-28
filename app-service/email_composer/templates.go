package email_composer

import (
	"fmt"

	"github.com/matcornic/hermes/v2"
)

type PasswordReset struct {
	RecipientName string
	CompanyName   string
	ResetURL      string
}

func NewPasswordReset(recipientName string, resetUrl string) *PasswordReset {
	return &PasswordReset{
		RecipientName: recipientName,
		CompanyName:   configObject.Email_CompanyName,
		ResetURL:      resetUrl,
	}
}

func (r *PasswordReset) Name() string {
	return "PasswordReset"
}

func (r *PasswordReset) Email() hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: r.RecipientName,
			Intros: []string{
				fmt.Sprintf("You have received this email because a password reset request for %s account was received.", r.CompanyName),
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to reset your password:",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Reset your password",
						Link:  r.ResetURL,
					},
				},
			},
			Outros: []string{
				"If you did not request a password reset, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}
