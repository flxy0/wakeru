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
