package email_composer

import (
	"fmt"

	"github.com/matcornic/hermes/v2"

	"sonamusica-backend/config"
)

var (
	configObject = config.Get()
)

// NewComposer receives no input parameters and takes the values directly from config.
// We assume that these values are app-level, so no need to be configurable via input parameters.
func NewComposer() *hermes.Hermes {
	return &hermes.Hermes{
		Product: hermes.Product{
			// Appears in header & footer of e-mails
			Name:      configObject.Email_CompanyName,
			Link:      configObject.Email_BaseAppURL,
			Logo:      configObject.LogoURL,
			Copyright: fmt.Sprintf("Copyright © 2023 %s. All rights reserved", configObject.Email_CompanyName),
		},
	}
}
