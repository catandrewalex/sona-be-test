package logging

import "sonamusica-backend/config"

var (
	configObject     = config.Get()
	HTTPServerLogger = NewGoLogger("HTTPServer", GetLevel(configObject.LogLevel))
	AppLogger        = NewGoLogger("App", GetLevel(configObject.LogLevel))
)
