package logging

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Trace(format string, a ...interface{})
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Error(format string, a ...interface{})
	Fatal(format string, a ...interface{})
	Panic(format string, a ...interface{})
}

type LogLevel int

func (l LogLevel) String() string {
	return logLevelToStringMap[l]
}

func GetLevel(logLevelStr string) LogLevel {
	return stringToLogLevelMap[logLevelStr]
}

const (
	LogLevel_None  LogLevel = iota
	LogLevel_Trace LogLevel = 1
	LogLevel_Debug LogLevel = 2
	LogLevel_Info  LogLevel = 3
	LogLevel_Warn  LogLevel = 4
	LogLevel_Error LogLevel = 5
	LogLevel_Fatal LogLevel = 6
)

var (
	logLevelToStringMap = map[LogLevel]string{
		LogLevel_None:  "NONE",
		LogLevel_Trace: "TRACE",
		LogLevel_Debug: "DEBUG",
		LogLevel_Info:  "INFO",
		LogLevel_Warn:  "WARN",
		LogLevel_Error: "ERROR",
		LogLevel_Fatal: "FATAL",
	}

	stringToLogLevelMap = func() map[string]LogLevel {
		result := make(map[string]LogLevel, 0)
		for k, v := range logLevelToStringMap {
			result[v] = k
		}
		return result
	}()
)

var (
	DefaultGoLog = log.New(os.Stdout, "", log.Ldate|log.LUTC|log.Ltime|log.Lshortfile)
)

func NewGoLogger(name string, minLogLevel LogLevel) Logger {
	return newGoLogger(DefaultGoLog, name, minLogLevel)
}

type goLogger struct {
	log         *log.Logger
	name        string
	minLogLevel LogLevel
	callDepth   int
}

const (
	DefaultCallDepth int = 2 // this is used by Go's log.Output. TLDR; this value must equal 2 to be able to print the short file name + line
)

func newGoLogger(log *log.Logger, name string, minLogLevel LogLevel) Logger {
	fmt.Printf("Constructing new goLogger: name='%s' minLogLevel='%v'\n", name, minLogLevel)
	return &goLogger{
		log:         log,
		name:        name,
		minLogLevel: minLogLevel,
		callDepth:   DefaultCallDepth,
	}
}

func (l *goLogger) Trace(format string, a ...interface{}) {
	if l.minLogLevel <= LogLevel_Trace {
		l.log.Output(l.callDepth, fmt.Sprintf("[%s] [TRACE] %s", l.name, fmt.Sprintf(format, a...)))
	}
}

func (l *goLogger) Debug(format string, a ...interface{}) {
	if l.minLogLevel <= LogLevel_Debug {
		l.log.Output(l.callDepth, fmt.Sprintf("[%s] [DEBUG] %s", l.name, fmt.Sprintf(format, a...)))
	}
}

func (l *goLogger) Info(format string, a ...interface{}) {
	if l.minLogLevel <= LogLevel_Info {
		l.log.Output(l.callDepth, fmt.Sprintf("[%s] [INFO] %s", l.name, fmt.Sprintf(format, a...)))
	}
}

func (l *goLogger) Warn(format string, a ...interface{}) {
	if l.minLogLevel <= LogLevel_Warn {
		l.log.Output(l.callDepth, fmt.Sprintf("[%s] [WARN] %s", l.name, fmt.Sprintf(format, a...)))
	}
}

func (l *goLogger) Error(format string, a ...interface{}) {
	if l.minLogLevel <= LogLevel_Error {
		l.log.Output(l.callDepth, fmt.Sprintf("[%s] [ERROR] %s", l.name, fmt.Sprintf(format, a...)))
	}
}

func (l *goLogger) Fatal(format string, a ...interface{}) {
	if l.minLogLevel <= LogLevel_Fatal {
		l.log.Output(l.callDepth, fmt.Sprintf("[%s] [FATAL] %s", l.name, fmt.Sprintf(format, a...)))
	}
}

func (l *goLogger) Panic(format string, a ...interface{}) {
	if l.minLogLevel <= LogLevel_Fatal {
		s := fmt.Sprintf(format, a...)
		l.log.Output(l.callDepth, fmt.Sprintf("[%s] [FATAL] %s", l.name, s))
		panic(s)
	}
}
