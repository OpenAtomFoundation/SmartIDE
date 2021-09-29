package common

import (
	"fmt"
	"log"
	"os"
)

//
type smartIDELogStruct struct {
}

var SmartIDELog *smartIDELogStruct

func (sLog *smartIDELogStruct) Error(err interface{}) (reErr error) {

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		log.Fatal(err)
		os.Exit(1)
	}

	return nil
}

func (sLog *smartIDELogStruct) Fatal(fatal error) (reErr error) {
	//TODO: 日志写入到文件中时，期望的效果是：msg + stack + log file path
	if fatal != nil {
		fmt.Fprintln(os.Stderr, "Error:", fatal)
		log.Fatal(fatal)
		os.Exit(1)
	}

	return nil
}

func (sLog *smartIDELogStruct) Info(info string) (err error) {

	log.Println(info)

	return nil
}

func (sLog *smartIDELogStruct) Debug(info string) (err error) {
	log.Println(info)
	return nil
}

func (sLog *smartIDELogStruct) Warning(warning string) (err error) {
	log.Println(warning)
	return nil
}
