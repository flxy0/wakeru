package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/flxy0/wakeru/config"
	"github.com/flxy0/wakeru/files"
	"github.com/flxy0/wakeru/hashgen"
	"github.com/flxy0/wakeru/helpers"
)

// Generate a copy of the templates dir in helplers/ for compiling.
//go:generate rm -rf ./helpers/templates
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

	helpers.LogErr(err)

	data := struct {
		InstanceName string
		AllowGenPage bool
		Error        string
	}{
		InstanceName: config.ConfigInstanceName,
		AllowGenPage: config.ConfigAllowGeneration,
		Error:        "",
	}

	tmplErr := indexTmplRender.Execute(w, data)
	helpers.LogErr(tmplErr)
}

func embedFsSub() http.FileSystem {
	subFs, err := fs.Sub(cssFile, "templates")
	helpers.LogErr(err)
	return http.FS(subFs)
}

func genRedirect(w http.ResponseWriter, r *http.Request) {
	// In case the Generate is set to false in the config, navigating to /generate manually will redirect back to the main page.
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

func handleArgs() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			handleInitArg()
		}
	}
}

func handleInitArg() {
	config.GenerateDefaultConfigInCurrentDir()
	fmt.Println("config created in current directory!")
	fmt.Println("feel free to edit the parameters and start the application by omiting the init")
	// Exit application because init is not supposed to run the server right away.
	os.Exit(0)
}

func main() {
	handleArgs()
	config.ParseConfig()

	mux := http.NewServeMux()

	// index route
	go mux.HandleFunc("/", renderIndexPage)

	// hash generation related routes
	// can be disabled by setting Generate to false in config.toml
	if config.ConfigAllowGeneration {
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
	go mux.Handle("/style.css", http.FileServer(embedFsSub()))

	fmt.Println("----- server starting -----")
	fmt.Println("now running app on http://localhost:5050")
	http.ListenAndServe(":5050", mux)
}
