package util

import (
	"sonamusica-backend/config"
	"sonamusica-backend/logging"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("UtilityMethods", logging.GetLevel(configObject.LogLevel))
)
