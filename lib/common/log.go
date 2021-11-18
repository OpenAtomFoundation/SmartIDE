package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gookit/color"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//
type smartIDELogStruct struct {
}

var SmartIDELog *smartIDELogStruct

var isDebugLevel bool = false

func (sLog *smartIDELogStruct) InitLogger(logLevel string) {
	if strings.ToLower(strings.TrimSpace(logLevel)) == "debug" {
		isDebugLevel = true
	}

	initLogger()
}

//
func (sLog *smartIDELogStruct) Error(err interface{}, headers ...string) (reErr error) {

	if err != nil {
		headers = RemoveDuplicatesAndEmpty(headers)
		msg := fmt.Sprint(strings.Join(headers, " "), err)

		prefix := getPrefix(zapcore.ErrorLevel)
		stack := string(debug.Stack())
		if isDebugLevel {
			fmt.Println(strings.Join([]string{prefix, msg, stack}, "; "))
		} else {
			fmt.Println(strings.Join([]string{prefix, msg}, "; "))
		}

		sugarLogger.Error(msg, stack)
		os.Exit(1)
	}

	return nil
}

//
func (sLog *smartIDELogStruct) Fatal(fatal interface{}, headers ...string) (reErr error) {
	if fatal != nil {
		headers = RemoveDuplicatesAndEmpty(headers)
		msg := fmt.Sprint(strings.Join(headers, " "), fatal)

		prefix := getPrefix(zapcore.FatalLevel)
		stack := string(debug.Stack())
		fmt.Println(strings.Join([]string{prefix, msg, stack}, "; "))

		sugarLogger.Fatal(strings.Join([]string{prefix, msg, stack}, "; "))
		os.Exit(1)
	}

	return nil
}

//
func (sLog *smartIDELogStruct) Info(info ...string) (err error) {
	if len(info) <= 0 {
		return nil
	}

	info = RemoveDuplicatesAndEmpty(info)
	msg := strings.Join(info, " ")

	prefix := getPrefix(zapcore.InfoLevel)
	fmt.Println(prefix, msg)

	sugarLogger.Info(msg)

	return nil
}

//
func (sLog *smartIDELogStruct) InfoF(format string, args ...interface{}) (err error) {

	validF(format, args...)

	msg := fmt.Sprintf(format, args...)

	return SmartIDELog.Info(msg)
}

func validF(format string, args ...interface{}) {
	if strings.Count(format, "%v") != len(args) {
		msg := fmt.Sprintf(": format tag reads arg count != length of args; format: %v, values: %v", format, args)
		panic(msg)
	}
}

//
func (sLog *smartIDELogStruct) DebugF(format string, args ...interface{}) (err error) {

	validF(format, args...)

	msg := fmt.Sprintf(format, args...)

	return SmartIDELog.Debug(msg)
}

func (sLog *smartIDELogStruct) Debug(info ...string) (err error) {

	if len(info) <= 0 {
		return nil
	}

	info = RemoveDuplicatesAndEmpty(info)
	msg := strings.Join(info, " ")

	prefix := getPrefix(zapcore.DebugLevel)
	if isDebugLevel {
		fmt.Println(prefix, msg)
	}

	sugarLogger.Debug(msg)

	return nil
}

// 输出到控制台，但是不加任何的修饰
func (sLog *smartIDELogStruct) Console(info ...interface{}) (err error) {
	if len(info) <= 0 {
		return nil
	}

	fmt.Println(info...)
	sugarLogger.Info(info...)

	return nil
}

//
func (sLog *smartIDELogStruct) Warning(warning ...string) (err error) {

	msg := strings.Join(warning, " ")
	if len(msg) <= 0 {
		return nil
	}

	prefix := getPrefix(zapcore.WarnLevel)
	if isDebugLevel {
		fmt.Println(prefix, msg)
	}

	sugarLogger.Warn(msg)

	return nil
}

//
func (sLog *smartIDELogStruct) WarningF(format string, args ...interface{}) (err error) {

	validF(format, args...)

	msg := fmt.Sprintf(format, args...)
	return SmartIDELog.Warning(msg)
}

//var logger *zap.Logger
var sugarLogger *zap.SugaredLogger

func initLogger() {
	// https://developpaper.com/use-of-golang-log-zap/
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel),
	)

	logger := zap.New(core, zap.AddCaller())
	sugarLogger = logger.Sugar()
}

//
func getEncoder() zapcore.Encoder {

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "file",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,                            // 日志级别大写
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), // 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	/*   zapConfig :=  zap.Config{
	     Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
	     Development:       false,
	     DisableCaller:     false,
	     DisableStacktrace: false,
	     Sampling:          nil,
	     Encoding:          "json",
	     EncoderConfig: zapcore.EncoderConfig{
	         MessageKey:     "msg",
	         LevelKey:       "level",
	         TimeKey:        "time",
	         NameKey:        "logger",
	         CallerKey:      "file",
	         StacktraceKey:  "stacktrace",
	         LineEnding:     zapcore.DefaultLineEnding,
	         EncodeLevel:    zapcore.LowercaseLevelEncoder,
	         EncodeTime:     zapcore.ISO8601TimeEncoder,
	         EncodeDuration: zapcore.SecondsDurationEncoder,
	         EncodeCaller:   zapcore.ShortCallerEncoder,
	         EncodeName:     zapcore.FullNameEncoder,
	     },
	     OutputPaths:      []string{"/tmp/zap.log"},
	     ErrorOutputPaths: []string{"/tmp/zap.log"},
	     InitialFields: map[string]interface{}{
	         "app": "test",
	     },
	 } */

	return zapcore.NewConsoleEncoder(encoderConfig)
}

//
func getLogWriter() zapcore.WriteSyncer {

	// common.IsLaunchedByDebugger()
	dirname, err := os.UserHomeDir() // home dir
	if err != nil {
		log.Fatal(err)
	}
	t := time.Now()
	logFileName := fmt.Sprintf("%v.log", t.Format("20060102"))
	logFilePath := filepath.Join(dirname, ".ide", logFileName) // current user dir + ...

	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFilePath, // ⽇志⽂件路径
		MaxSize:    12,          // 单位为MB,默认为512MB
		MaxBackups: 30,          // 备份的文件最大的数量
		LocalTime:  true,        // 本地时间
		MaxAge:     7,           // 文件最多保存多少天
		Compress:   false,       // 是否压缩
	}
	return zapcore.AddSync(lumberJackLogger)
}

//
func getPrefix(logLevel zapcore.Level) string {
	t := time.Now()
	timeStr := t.Format("2006-01-02 15:04:05.000")

	levelStr := ""
	switch logLevel {
	case zapcore.ErrorLevel:
		levelStr = color.Error.Sprint("ERROR") //levelStr = fmt.Sprintf("\x1b[31;1m%v\x1b[0m", "ERROR")
	case zapcore.InfoLevel:
		levelStr = color.Info.Sprint("INFO") // fmt.Sprintf("\x1b[34;1m%s\x1b[0m", "INFO")
	case zapcore.FatalLevel:
		levelStr = color.BgRed.Sprint("FATAL") //fmt.Sprintf("\x1b[31;1m%s\x1b[0m", "FATAL")
	case zapcore.PanicLevel:
		levelStr = color.BgLightRed.Sprint("FATAL")
	case zapcore.WarnLevel:
		levelStr = color.Warn.Sprint("WARNING") //levelStr = fmt.Sprintf("\x1b[36;1m%s\x1b[0m", "WARNING")
	case zapcore.DebugLevel:
		levelStr = color.Debug.Sprint("DEBUG") //levelStr = fmt.Sprintf("\x1b[34;1m%s\x1b[0m", "DEBUG")
	}

	return fmt.Sprintf("%v %v ", timeStr, levelStr)
}
