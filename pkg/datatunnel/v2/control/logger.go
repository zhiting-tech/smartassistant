package control

type Logger interface {
	Tracef(format string, args ...interface{})

	Debugf(format string, args ...interface{})

	Infof(format string, args ...interface{})

	Printf(format string, args ...interface{})

	Warnf(format string, args ...interface{})

	Warningf(format string, args ...interface{})

	Errorf(format string, args ...interface{})

	Fatalf(format string, args ...interface{})

	Panicf(format string, args ...interface{})

	Trace(args ...interface{})

	Debug(args ...interface{})

	Info(args ...interface{})

	Print(args ...interface{})

	Warn(args ...interface{})

	Warning(args ...interface{})

	Error(args ...interface{})

	Fatal(args ...interface{})

	Panic(args ...interface{})
}
