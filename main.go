package main

import (
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/flxy0/wakeru/files"
	"github.com/flxy0/wakeru/hashgen"
	"github.com/flxy0/wakeru/helpers"
)

// Type for Template Files
type templateFile struct {
	name     string
	contents string
}

// With the host being able to disable the hash generation route and functionality, we need to parse the argvs and not render the <a> element in the nav if the generation is disabled.
func renderIndexPage(w http.ResponseWriter, r *http.Request) {
	indexTmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/index.gohtml"))

	data := struct {
		DisableGenPage bool
		Error          string
	}{
		DisableGenPage: helpers.NoGenArgPassed(),
		Error:          "",
	}

	tmplErr := indexTmpl.Execute(w, data)
	if tmplErr != nil {
		log.Println(tmplErr)
	}
}

// In case the "-nogen" flag is passed, navigating to /generate manually will redirect back to the main page.
func genRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// This function handles serving static files such as CSS or images the host wants them.
func renderStaticFile(filepath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath)
	}
}

// This function takes care of all the content serving
// It checks whether the URL has the first 20 digits of an existing string
// and whether the file exists in the corresponding full hash named directory
func handleServeContent(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	userHash := urlParts[2]
	filename := urlParts[3]
	var dirMatch string

	for _, v := range helpers.ServeDirs {
		if userHash == v[:20] {
			dirMatch = v
			break
		}
	}

	// If dirMatch doesn't get a value assigned to it, the hash doesn't exist so we need to inform the user
	if dirMatch == "" {
		log.Println(w, "Hash is wrong or doesn't exist")
		fmt.Fprintf(w, "Hash is wrong or doesn't exist")
		return
	}

	filePath := fmt.Sprintf("uploads/%s/%s", dirMatch, filename)

	fmt.Println("." + (strings.Split(filename, ".")[1]))
	fmt.Println(mime.TypeByExtension("." + (strings.Split(filename, ".")[1])))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(w, "File does not exist")
		return
	}

	http.ServeFile(w, r, filePath)
}

func main() {
	mux := http.NewServeMux()

	// index route
	go mux.HandleFunc("/", renderIndexPage)

	// hash generation related routes
	// can be disabled by passing `-nogen` arg
	if !helpers.NoGenArgPassed() {
		go mux.HandleFunc("/generate", hashgen.Generate)
	} else {
		go mux.HandleFunc("/generate", genRedirect)
	}

	// upload form and handling of post errors
	go mux.HandleFunc("/upload", files.Upload)

	// serve uploads
	go mux.HandleFunc("/uploads/", handleServeContent)

	// view files corresponding to hash
	go mux.HandleFunc("/viewfiles/", files.ViewFiles)

	// serve static file(s) if need be
	go mux.HandleFunc("/style.css", renderStaticFile("templates/style.css"))

	fmt.Println("now running app on http://localhost:5050")
	http.ListenAndServe(":5050", mux)
}
