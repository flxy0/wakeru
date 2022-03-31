package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/flxy0/wakeru/files"
	"github.com/flxy0/wakeru/hashgen"
	"github.com/flxy0/wakeru/helpers"
)

// Generate a copy of the templates dir for compiling.
//go:generate cp -r ./templates/ ./helpers/templates

// Type for Template Files
type templateFile struct {
	name     string
	contents string
}

//go:embed templates/style.css
var cssFile embed.FS

func renderIndexPage(w http.ResponseWriter, r *http.Request) {
	// With the host being able to disable the hash generation route and functionality, we need to parse the argvs and not render the <a> element in the nav if the generation is disabled.

	// Two step process due to go:embed
	baseTmplRender := template.Must(template.ParseFS(helpers.TemplateDir, "templates/base.gohtml"))
	indexTmplRender, err := template.Must(baseTmplRender.Clone()).ParseFS(helpers.TemplateDir, "templates/index.gohtml")

	if err != nil {
		log.Println(err)
	}

	data := struct {
		DisableGenPage bool
		Error          string
	}{
		DisableGenPage: helpers.NoGenArgPassed(),
		Error:          "",
	}

	tmplErr := indexTmplRender.Execute(w, data)
	if tmplErr != nil {
		log.Println(tmplErr)
	}
}

func genRedirect(w http.ResponseWriter, r *http.Request) {
	// In case the "-nogen" flag is passed, navigating to /generate manually will redirect back to the main page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleServeContent(w http.ResponseWriter, r *http.Request) {
	// This function takes care of all the content serving
	// It checks whether the URL has the first 20 digits of an existing string
	// and whether the file exists in the corresponding full hash named directory

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
		return
	}

	filePath := fmt.Sprintf("uploads/%s/%s", dirMatch, filename)

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
	go mux.Handle("/style.css", http.FileServer(http.FS(cssFile)))

	fmt.Println("now running app on http://localhost:5050")
	http.ListenAndServe(":5050", mux)
}
