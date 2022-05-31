// Package logger 提供日志记录功能
// 每条日志均会输出 app、module 字段，分别代表应用，模块，默认值为 default，此两个字段会合并为 fluentd 的 tag
// 设备插件 app 请使用 plugin_{插件名称} 的格式，方便后台查看
// debug 模式下会输出 caller、package 字段作为跟踪
package logger

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	NewEntrySkip    = 4 // 外部通过new新的logger，获取的caller层级
	DirectEntrySkip = 5 // 直接调用封装的方法，获取的caller层级
)

var (
	newLogger = logger()
	entry     = newEntry()
	fields    logrus.Fields
	option    Option
)

type Option struct {
	Level     logrus.Level
	Output    io.Writer
	LogFormat logrus.Formatter
	Debug     bool
}

type Entry struct {
	Entry *logrus.Entry
	Skip  int // caller层级
}

// generate a global instance
func logger() *logrus.Logger {
	return logrus.New()
}

// New 外部使用
func New() *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(option.LogFormat)
	l.SetLevel(option.Level)
	l.SetOutput(option.Output)
	l.WithFields(fields)

	return l
}

// InitLogger generate a global instance
func InitLogger(output io.Writer, level logrus.Level, f logrus.Fields, debug bool, hooks ...logrus.Hook) {
	output = io.MultiWriter(output, getRemoteWriter())
	newLogger.SetOutput(output)
	newLogger.SetLevel(level)

	format := &logrus.JSONFormatter{}
	newLogger.SetFormatter(format)

	for _, h := range hooks {
		if h == nil {
			continue
		}
		newLogger.Hooks.Add(h)
	}
	if _, ok := f["app"]; !ok {
		f["app"] = "default"
	}
	if _, ok := f["module"]; !ok {
		f["module"] = "default"
	}
	fields = f
	option = Option{level, output, format, debug}

	// 初始化Entry
	entry = newEntry()
}

// 内部初始化
func newEntry() *Entry {
	return &Entry{
		Entry: newLogger.WithFields(fields),
		Skip:  DirectEntrySkip,
	}
}

// NewEntry 外部初始化
func NewEntry() *Entry {
	return &Entry{
		Entry: newLogger.WithFields(fields),
		Skip:  NewEntrySkip,
	}
}

func SetDebug(IsDebug bool) {
	option.Debug = IsDebug
}

func Debug(args ...interface{}) {
	entry.Debug(args...)
}

func Info(args ...interface{}) {
	entry.Info(args...)
}

func Warning(args ...interface{}) {
	entry.Warning(args...)
}

func Warn(args ...interface{}) {
	entry.Warn(args...)
}

func Error(args ...interface{}) {
	entry.Errorln(args...)
}

func Panic(args ...interface{}) {
	entry.Panic(args...)
}

func Debugf(format string, args ...interface{}) {
	entry.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	entry.Infof(format, args...)
}

func Warningf(format string, args ...interface{}) {
	entry.Warningf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	entry.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	entry.Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	entry.Panicf(format, args...)
}

func Panicln(args ...interface{}) {
	entry.Panicln(args...)
}

func Printf(format string, args ...interface{}) {
	entry.Printf(format, args...)
}

func Println(args ...interface{}) {
	entry.Println(args...)
}

func (entry *Entry) withCaller() *Entry {
	caller := getCallerInfo(entry.Skip)
	pkg := getPackageName(caller)

	appendFields := logrus.Fields{"caller": caller, "package": pkg}
	entry.WithFields(appendFields)

	return entry
}

func (entry *Entry) WithField(key string, value interface{}) *Entry {
	entry.Entry = entry.Entry.WithField(key, value)
	return entry
}

func (entry *Entry) WithFields(field logrus.Fields) *Entry {
	entry.Entry = entry.Entry.WithFields(field)
	return entry
}

func (entry *Entry) Log(level logrus.Level, args ...interface{}) {
	if option.Debug {
		entry.withCaller()
	}
	entry.Entry.Log(level, args...)
}

func (entry *Entry) Logf(level logrus.Level, format string, args ...interface{}) {
	if option.Debug {
		entry.withCaller()
	}
	entry.Entry.Logf(level, format, args...)
}

func (entry *Entry) Logln(level logrus.Level, args ...interface{}) {
	if option.Debug {
		entry.withCaller()
	}
	entry.Entry.Logln(level, args...)
}

func (entry *Entry) Trace(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.TraceLevel) {
		entry.Log(logrus.TraceLevel, args...)
	}
}

func (entry *Entry) Debug(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.DebugLevel) {
		entry.Log(logrus.DebugLevel, args...)
	}
}

func (entry *Entry) Info(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.InfoLevel) {
		entry.Log(logrus.InfoLevel, args...)
	}
}

func (entry *Entry) Warn(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.WarnLevel) {
		entry.Log(logrus.WarnLevel, args...)
	}
}

func (entry *Entry) Warning(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.WarnLevel) {
		entry.Log(logrus.WarnLevel, args...)
	}
}

func (entry *Entry) Error(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.ErrorLevel) {
		entry.Log(logrus.ErrorLevel, args...)
	}
}

func (entry *Entry) Fatal(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.FatalLevel) {
		entry.Log(logrus.FatalLevel, args...)
	}
}

func (entry *Entry) Panic(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.PanicLevel) {
		entry.Log(logrus.PanicLevel, args...)
	}
}

func (entry *Entry) Tracef(format string, args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.TraceLevel) {
		entry.Logf(logrus.TraceLevel, format, args...)
	}
}

func (entry *Entry) Debugf(format string, args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.DebugLevel) {
		entry.Logf(logrus.DebugLevel, format, args...)
	}
}

func (entry *Entry) Infof(format string, args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.InfoLevel) {
		entry.Logf(logrus.InfoLevel, format, args...)
	}
}

func (entry *Entry) Warnf(format string, args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.WarnLevel) {
		entry.Logf(logrus.WarnLevel, format, args...)
	}
}

func (entry *Entry) Warningf(format string, args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.WarnLevel) {
		entry.Logf(logrus.WarnLevel, format, args...)
	}
}

func (entry *Entry) Errorf(format string, args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.ErrorLevel) {
		entry.Logf(logrus.ErrorLevel, format, args...)
	}
}

func (entry *Entry) Fatalf(format string, args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.FatalLevel) {
		entry.Logf(logrus.FatalLevel, format, args...)
	}
}

func (entry *Entry) Panicf(format string, args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.PanicLevel) {
		entry.Logf(logrus.PanicLevel, format, args...)
	}
}

func (entry *Entry) Traceln(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.TraceLevel) {
		entry.Logln(logrus.TraceLevel, args...)
	}
}

func (entry *Entry) Debugln(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.DebugLevel) {
		entry.Logln(logrus.DebugLevel, args...)
	}
}

func (entry *Entry) Infoln(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.InfoLevel) {
		entry.Logln(logrus.InfoLevel, args...)
	}
}

func (entry *Entry) Warnln(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.WarnLevel) {
		entry.Logln(logrus.WarnLevel, args...)
	}
}

func (entry *Entry) Warningln(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.WarnLevel) {
		entry.Logln(logrus.WarnLevel, args...)
	}
}

func (entry *Entry) Errorln(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.ErrorLevel) {
		entry.Logln(logrus.ErrorLevel, args...)
	}
}

func (entry *Entry) Fatalln(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.FatalLevel) {
		entry.Logln(logrus.FatalLevel, args...)
	}
}

func (entry *Entry) Panicln(args ...interface{}) {
	if newLogger.IsLevelEnabled(logrus.PanicLevel) {
		entry.Logln(logrus.PanicLevel, args...)
	}
}

func (entry *Entry) Print(args ...interface{}) {
	entry.Log(logrus.InfoLevel, args...)
}

func (entry *Entry) Printf(format string, args ...interface{}) {
	entry.Logf(logrus.InfoLevel, format, args...)
}

func (entry *Entry) Println(args ...interface{}) {
	entry.Logln(logrus.InfoLevel, args...)
}

func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<???>"
		line = 0
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func getPackageName(caller string) string {
	index := strings.LastIndex(caller, "/")
	tmp := caller[0:index]
	index = strings.LastIndex(tmp, "/")
	return tmp[index+1:]
}
