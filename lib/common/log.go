package common

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

//
type smartIDELogStruct struct {
}

var SmartIDELog *smartIDELogStruct

//
func (sLog *smartIDELogStruct) Error(err interface{}, headers ...string) (reErr error) {

	if err != nil {
		msg := fmt.Sprint("[Error]", strings.Join(headers, " "), err)
		stack := string(debug.Stack())
		fmt.Fprintln(os.Stderr, msg)
		log.Fatal(msg, stack)
		os.Exit(1)
	}

	return nil
}

//
func (sLog *smartIDELogStruct) Fatal(fatal interface{}, headers ...string) (reErr error) {
	//TODO: 日志写入到文件中时，期望的效果是：msg + stack + log file path
	if fatal != nil {
		msg := fmt.Sprint("[Fatal]", strings.Join(headers, " "), fatal)
		stack := string(debug.Stack())
		fmt.Fprintln(os.Stderr, msg, stack)
		log.Fatal(msg, stack)
		os.Exit(1)
	}

	return nil
}

//
func (sLog *smartIDELogStruct) Info(info ...string) (err error) {
	log.Println("[Info]", strings.Join(info, " "))

	return nil
}

//
func (sLog *smartIDELogStruct) Debug(info ...string) (err error) {
	log.Println("[Debug]", strings.Join(info, " "))

	return nil
}

//
func (sLog *smartIDELogStruct) Warning(warning ...string) (err error) {
	log.Println("[Warning]", strings.Join(warning, " "))
	fmt.Println("[Warning]", strings.Join(warning, " "))
	return nil
}
