package helpers

import (
	"embed"
	"io/ioutil"
	"log"
)

//go:embed templates/*.gohtml
var TemplateDir embed.FS

// Global variable that stores a slice with all the existing directories
var ServeDirs = FetchDirList()

func FetchDirList() []string {
	// Simple helper function to fetch a splice with existing directories
	// Mainly used to assign to the serveDirs global variable
	dirArr, err := ioutil.ReadDir("uploads/")

	if err != nil {
		log.Println(err)
		panic("No uploads folder found")
	}

	dirs := make([]string, len(dirArr))
	for i, v := range dirArr {
		dirs[i] = v.Name()
	}

	return dirs
}

func LogErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func LogErrorAndAppendToErrList(err error, errStr string, errList []string) ([]string, bool) {
	additErr := false
	if err != nil {
		log.Println(err)
		errList = append(errList, errStr)
		additErr = true
	} else if errStr != "" {
		errList = append(errList, errStr)
	}
	return errList, additErr
}

func LogErrAndReturnErrString(err error, errStr string) string {
	if err != nil {
		log.Println(err)
		return errStr
	} else {
		return ""
	}
}
