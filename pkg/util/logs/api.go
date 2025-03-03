package logs

func init() {
	defaultLogger = NewStdLogger(LevelInfo)
}

func SetLogger(logger Logger) {
	defaultLogger = logger
}

const (
	LevelDebug = 1 + iota
	LevelInfo
	LevelWarn
	LevelError
)

// TODO: merge with crafter logger package
type Logger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Flush()
	SetLevel(level int) error
}

var defaultLogger Logger

func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

func Error(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

func Warn(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

func Info(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

func Debug(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

func Flush() {
	defaultLogger.Flush()
}

func SetLevel(level int) {
	defaultLogger.SetLevel(level)
}
