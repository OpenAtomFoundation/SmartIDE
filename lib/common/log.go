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

	if fatal != nil {
		fmt.Fprintln(os.Stderr, "Error:", fatal)
		log.Fatal(fatal)
		os.Exit(1)
	}

	return nil
}

func (sLog *smartIDELogStruct) Info(info string) (err error) {
	//TOOD:

	return nil
}

func (sLog *smartIDELogStruct) Debug(info string) (err error) {
	//TOOD:

	return nil
}
