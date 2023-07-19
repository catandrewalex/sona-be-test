package util

import (
	"sonamusica-backend/config"
	"sonamusica-backend/logging"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("Utility Methods", logging.GetLevel(configObject.LogLevel))
)
