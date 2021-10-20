package common

import (
	"fmt"
	"log"
	"os"
	"strings"
)

//
type smartIDELogStruct struct {
}

var SmartIDELog *smartIDELogStruct

func (sLog *smartIDELogStruct) Error(err interface{}, headers ...string) (reErr error) {

	if err != nil {
		msg := fmt.Sprint("Error: ", strings.Join(headers, " "), err)
		fmt.Fprintln(os.Stderr, msg)
		log.Fatal(msg)
		os.Exit(1)
	}

	return nil
}

func (sLog *smartIDELogStruct) Fatal(fatal interface{}, headers ...string) (reErr error) {
	//TODO: 日志写入到文件中时，期望的效果是：msg + stack + log file path
	if fatal != nil {
		msg := fmt.Sprint("Fatal: ", strings.Join(headers, " "), fatal)
		fmt.Fprintln(os.Stderr, msg)
		log.Fatal(msg)
		os.Exit(1)
	}

	return nil
}

func (sLog *smartIDELogStruct) Info(info ...string) (err error) {
	log.Println(strings.Join(info, " "))

	return nil
}

func (sLog *smartIDELogStruct) Debug(info ...string) (err error) {
	log.Println(strings.Join(info, " "))

	return nil
}

func (sLog *smartIDELogStruct) Warning(warning ...string) (err error) {
	log.Println(strings.Join(warning, " "))
	fmt.Println(strings.Join(warning, " "))
	return nil
}
