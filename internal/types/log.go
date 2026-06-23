package types

type ILog interface {
	Start() func()
	LogTrace(format string, v ...any)
	LogDebug(format string, v ...any)
	LogInfo(format string, v ...any)
	LogWarn(format string, v ...any)
	LogError(err error)
	LogFatal(err error)
	LogPanic(err error)
}

const (
	StrTrace = "TRACE"
	StrDebug = "DEBUG"
	StrInfo  = "INFO"
	StrWarn  = "WARN"
	StrError = "ERROR"
	StrPanic = "PANIC"
	StrFatal = "FATAL"
	StrNoLog = "NOLOG"
)

type LogLevel int8

const (
	Trace LogLevel = iota
	Debug
	Info
	Warn
	Error
	Panic
	Fatal
	NoLog
)
