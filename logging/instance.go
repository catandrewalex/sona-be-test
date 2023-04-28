package logging

var (
	HTTPServerLogger = NewGoLogger("HTTPServer", LogLevel_Info)
	AppLogger        = NewGoLogger("App", LogLevel_Info)
)
