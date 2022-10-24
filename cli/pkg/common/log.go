/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-24 15:04:50
 */
package common

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/leansoftX/smartide-cli/internal/model"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type smartIDELogStruct struct {
	Ws_id      string
	ParentId   int
	TekEventId string
}

var (
	ServerToken    string
	ServerUserName string
	ServerUserGuid string
	ServerHost     string
	Mode           string
)

var SmartIDELog = &smartIDELogStruct{Ws_id: "", ParentId: 0}

// 进入诊断模式
var isDebugLevel bool = false

type lastMsg struct {
	Message    string
	CreateTime time.Time
	TimeOut    time.Duration
	Level      zapcore.Level
}

var _lastMsg lastMsg

func isRepeat(message string, level zapcore.Level) bool {
	repeat := false
	fmt.Sprintln(time.Since(_lastMsg.CreateTime))
	if (_lastMsg != lastMsg{}) && message != "" &&
		message == _lastMsg.Message && level == _lastMsg.Level {
		if _lastMsg.TimeOut < 0 || (_lastMsg.TimeOut >= 0 && time.Since(_lastMsg.CreateTime) <= _lastMsg.TimeOut) {
			repeat = true
		}

	}

	if !repeat {
		_lastMsg = lastMsg{Message: message, CreateTime: time.Now(), TimeOut: time.Minute * 10, Level: level}

	}

	return repeat
}

func (sLog *smartIDELogStruct) InitLogger(logLevel string) {
	if strings.ToLower(strings.TrimSpace(logLevel)) == "debug" {
		isDebugLevel = true
	}
	initLogger()
}

type entryptionKeyConfig struct {
	SecretKey     string
	IsReservePart bool
}

var _entryptionKeyConfigs []entryptionKeyConfig

// 添加需要加密的密钥信息
func (sLog *smartIDELogStruct) AddEntryptionKey(key string) {
	sLog.addEntryptionKey(key, false)
}

func (sLog *smartIDELogStruct) AddEntryptionKeyWithReservePart(key string) {
	sLog.addEntryptionKey(key, true)
}

func (sLog *smartIDELogStruct) addEntryptionKey(key string, isReservePart bool) {
	if strings.TrimSpace(key) == "" {
		return
	}
	isContain := false
	for _, item := range _entryptionKeyConfigs {
		if item.SecretKey == key {
			isContain = true
		}
	}
	if !isContain {
		_entryptionKeyConfigs = append(_entryptionKeyConfigs, entryptionKeyConfig{key, isReservePart})
	}
}

func entryptionKeys(contents []string) []string {
	result := []string{}
	for _, content := range contents {
		result = append(result, entryptionKey(content))
	}

	return result
}

func entryptionKey(content string) string {
	for _, entryptionKeyConfig := range _entryptionKeyConfigs {
		if entryptionKeyConfig.IsReservePart && len(entryptionKeyConfig.SecretKey) > 3 {
			newStr := entryptionKeyConfig.SecretKey[:3] + "******" + entryptionKeyConfig.SecretKey[len(entryptionKeyConfig.SecretKey)-3:]
			content = strings.ReplaceAll(content, entryptionKeyConfig.SecretKey, newStr)
		} else {
			content = strings.ReplaceAll(content, entryptionKeyConfig.SecretKey, "***")
		}
	}

	return content
}

func (sLog *smartIDELogStruct) Error(err interface{}, headers ...string) (reErr error) {

	if err == nil {
		return nil
	}

	// 基本的信息
	contents := headers
	contents = append(contents, fmt.Sprint(err))

	// 前缀
	prefix := getPrefix(zapcore.ErrorLevel)
	contents = RemoveDuplicatesAndEmpty(contents)

	// 堆栈
	stack := string(debug.Stack())
	fullContents := append(contents, stack)
	fullContents = RemoveDuplicatesAndEmpty(fullContents)
	if sLog.Ws_id != "" {
		go SendAndReceive("business", "workspaceLog", "", "", model.WorkspaceLog{
			Title:    "",
			ParentId: sLog.ParentId,
			Content:  strings.Join(fullContents, ";"),
			Ws_id:    sLog.Ws_id,
			Level:    4,
			Type:     1,
			StartAt:  time.Now(),
			EndAt:    time.Now(),
		})
	}
	fullContents = entryptionKeys(fullContents) // 加密密钥

	// 调试模式时向控制台输出堆栈
	if isDebugLevel {
		fmt.Println(prefix, strings.Join(fullContents, "; "))
	} else {
		fmt.Println(prefix, strings.Join(contents, "; "))
	}

	// 日志中一定输出完整的日志
	sugarLogger.Error(fullContents)
	os.Exit(1)

	return nil
}

func (sLog *smartIDELogStruct) Fatal(fatal interface{}, headers ...string) {
	if fatal != nil {
		// 基本信息
		contents := headers
		contents = append(contents, fmt.Sprint(fatal))

		// 前缀
		prefix := getPrefix(zapcore.FatalLevel)
		contents = append([]string{prefix}, contents...)

		// 堆栈
		stack := string(debug.Stack())
		contents = append(contents, stack)
		contents = RemoveDuplicatesAndEmpty(contents) // 去重
		contents = entryptionKeys(contents)           // 加密密钥

		// 打印日志
		fmt.Println(strings.Join(contents, "; "))

		// 记录日志
		sugarLogger.Fatal(strings.Join(contents, "; "))
		os.Exit(1)
	}
}

func (sLog *smartIDELogStruct) Info(args ...string) {
	args = RemoveDuplicatesAndEmpty(args)
	if len(args) <= 0 {
		return
	}

	msg := strings.Join(args, " ")
	if isRepeat(msg, zapcore.InfoLevel) { // 是否重复
		return
	}
	msg = entryptionKey(msg) // 加密

	prefix := getPrefix(zapcore.InfoLevel)
	fmt.Println(prefix, msg)
	if sLog.Ws_id != "" {
		go SendAndReceive("business", "workspaceLog", "", "", model.WorkspaceLog{
			Title:    "",
			ParentId: sLog.ParentId,
			Content:  msg,
			Ws_id:    sLog.Ws_id,
			Level:    1,
			Type:     1,
			StartAt:  time.Now(),
			EndAt:    time.Now(),
		})
	}
	sugarLogger.Info(msg)
}

func (sLog *smartIDELogStruct) InfoF(format string, args ...interface{}) {

	validF(format, args...)

	msg := fmt.Sprintf(format, args...)

	SmartIDELog.Info(msg)
}

func validF(format string, args ...interface{}) {
	if strings.Count(format, "%v") != len(args) {
		msg := fmt.Sprintf(": format tag reads arg count != length of args; format: %v, values: %v", format, args)
		panic(msg)
	}
}

func (sLog *smartIDELogStruct) DebugF(format string, args ...interface{}) {

	validF(format, args...)

	msg := fmt.Sprintf(format, args...)

	SmartIDELog.Debug(msg)
}

func (sLog *smartIDELogStruct) Debug(args ...string) {

	args = RemoveDuplicatesAndEmpty(args)
	if len(args) <= 0 {
		return
	}

	msg := strings.Join(args, " ")
	msg = entryptionKey(msg) // 加密

	prefix := getPrefix(zapcore.DebugLevel)
	if isDebugLevel {

		if isRepeat(msg, zapcore.DebugLevel) { // 是否重复
			return
		} else {
			fmt.Println(prefix, msg)
		}

	}
	if sLog.Ws_id != "" {
		go SendAndReceive("business", "workspaceLog", "", "", model.WorkspaceLog{
			Title:    "",
			ParentId: sLog.ParentId,
			Content:  msg,
			Ws_id:    sLog.Ws_id,
			Level:    3,
			Type:     1,
			StartAt:  time.Now(),
			EndAt:    time.Now(),
		})
	}
	sugarLogger.Debug(msg)
}

// 输出到控制台，但是不加任何的修饰
func (sLog *smartIDELogStruct) Console(args ...interface{}) {

	if len(args) <= 0 {
		return
	}

	strs := []string{}
	for _, item := range args {
		strs = append(strs, fmt.Sprint(item))
	}
	strs = entryptionKeys(strs)
	fmt.Println(strs)
	sugarLogger.Info(strs)
}

// 输出到控制台，但是不加任何的修饰
func (sLog *smartIDELogStruct) ConsoleDebug(args ...interface{}) {

	if len(args) <= 0 {
		return
	}

	sugarLogger.Info(args...)
}

// 输出到控制台，在一行
func (sLog *smartIDELogStruct) ConsoleInLine(args ...interface{}) {

	if len(args) <= 0 {
		return
	}
	strs := []string{}
	for _, item := range args {
		strs = append(strs, fmt.Sprint(item))
	}
	strs = entryptionKeys(strs) // 加密

	fmt.Printf("\r%v\r", strs)
	sugarLogger.Info(strs)
}

func (sLog *smartIDELogStruct) ImportanceWithError(err error) {
	if err == nil {
		return
	}
	if _, ok := err.(*exec.ExitError); !ok {
		sLog.Importance(err.Error())
	}

}

// 一些重要的信息，已warning的形式输出到控制台
func (sLog *smartIDELogStruct) Importance(infos ...string) {
	msg := strings.Join(infos, " ")
	if len(msg) <= 0 {
		return
	}

	if isRepeat(msg, zapcore.WarnLevel) { // 是否重复
		return
	}
	msg = entryptionKey(msg) // 加密

	prefix := getPrefix(zapcore.WarnLevel)
	fmt.Println(prefix, msg)

	sugarLogger.Warn(msg)

}

// 等待另外一个新日志
func (sLog *smartIDELogStruct) WaitingForAnother() {
	_lastMsg.TimeOut = -1
}

func (sLog *smartIDELogStruct) Warning(warning ...string) {

	msg := strings.Join(warning, " ")
	if len(msg) <= 0 {
		return
	}
	msg = entryptionKey(msg) // 加密

	prefix := getPrefix(zapcore.WarnLevel)
	if isDebugLevel {
		fmt.Println(prefix, msg)
	}
	if sLog.Ws_id != "" {
		go SendAndReceive("business", "workspaceLog", "", "", model.WorkspaceLog{
			Title:    "",
			ParentId: sLog.ParentId,
			Content:  msg,
			Ws_id:    sLog.Ws_id,
			Level:    2,
			Type:     1,
			StartAt:  time.Now(),
			EndAt:    time.Now(),
		})
	}
	sugarLogger.Warn(msg)
}

func (sLog *smartIDELogStruct) WarningF(format string, args ...interface{}) {

	validF(format, args...)

	msg := fmt.Sprintf(format, args...)
	SmartIDELog.Warning(msg)
}

// var logger *zap.Logger
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

func getLogWriter() zapcore.WriteSyncer {

	// common.IsLaunchedByDebugger()
	dirname, err := os.UserHomeDir() // home dir
	if err != nil {
		log.Fatal(err)
	}
	t := time.Now()
	logFileName := fmt.Sprintf("%v.log", t.Format("20060102"))
	logFilePath := PathJoin(dirname, ".ide", logFileName) // current user dir + ...

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
