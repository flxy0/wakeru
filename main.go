package main

import (
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"sr.ht/flxy/wakeru/files"
	"sr.ht/flxy/wakeru/hashgen"
	"sr.ht/flxy/wakeru/helpers"
)

// Type for Template Files
type templateFile struct {
	name     string
	contents string
}

// This function is responsible for rendering all the templates that don't need any extra logic
// or data input
func renderDatalessTemplate(tmpl template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error rendering a template")
		}
	}
}

func renderStaticFile(filepath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath)
	}
}

// This function handles the post form from the upload.gohtml
// It verifies that there's a valid hash and a file present upon submission
// If not, it will simply display a simple text message informing the user what is wrong

// This function takes care of all the content serving
// It checks whether the URL has the first 20 digits of an existing string
// and whether the file exists in the corresponding full hash named directory
func handleServeContent(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	userHash := urlParts[2]
	filename := urlParts[3]
	var dirMatch string

	for _, v := range helpers.ServeDirs {
		if strings.HasPrefix(v, userHash) && len(userHash) == 20 {
			dirMatch = v
		}
		// else {
		// 	fmt.Fprintf(w, "Hash is wrong or doesn't exist")
		// 	fmt.Println(userHash)
		// }
	}

	filePath := fmt.Sprintf("uploads/%s/%s", dirMatch, filename)

	fmt.Println("." + (strings.Split(filename, ".")[1]))
	fmt.Println(mime.TypeByExtension("." + (strings.Split(filename, ".")[1])))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(w, "File does not exist")
		return
	}

	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		log.Fatal(fileErr)
		return
	}

	fileStats, statErr := file.Stat()
	if statErr != nil {
		log.Fatal(statErr)
	}

	fileBuf := make([]byte, fileStats.Size())
	file.Read(fileBuf)

	w.Header().Add("Content-Type", mime.TypeByExtension("."+(strings.Split(filename, ".")[1])))
	w.Write(fileBuf)
	file.Close()
}

func main() {
	var (
		indexTmpl  = template.Must(template.ParseFiles("templates/base.gohtml", "templates/index.gohtml"))
		genTmpl    = template.Must(template.ParseFiles("templates/base.gohtml", "templates/gen.gohtml"))
		uploadTmpl = template.Must(template.ParseFiles("templates/base.gohtml", "templates/upload.gohtml"))
		// fileViewTmpl = template.Must(template.ParseFiles("templates/base.gohtml", "templates/view_files.gohtml"))
		// template_list    = []string{"templates/base.gothml", "templates/index.gohtml"}
		// templates =      template.Must(template.ParseFiles("templates/base.gohtml"))
		// index_tmpl, _ = template.Must(base_tmpl.Clone()).ParseFiles("templates/index.gohtml")
	)

	// template.Must()

	mux := http.NewServeMux()

	// index route
	go mux.HandleFunc("/", renderDatalessTemplate(*indexTmpl))

	// hash generation related routes
	go mux.HandleFunc("/generate", renderDatalessTemplate(*genTmpl))
	go mux.HandleFunc("/generated", hashgen.Generated)

	// upload form and handling of post errors
	go mux.HandleFunc("/upload", renderDatalessTemplate(*uploadTmpl))
	go mux.HandleFunc("/uploaded", files.Uploaded)

	// serve uploads
	go mux.HandleFunc("/uploads/", handleServeContent)

	// view files corresponding to hash
	// go mux.HandleFunc("/viewfiles", renderDatalessTemplate(*fileViewTmpl))
	go mux.HandleFunc("/viewfiles/", files.ViewFiles)
	// go mux.HandleFunc("/view_files/deletion", deleteFiles)

	// serve static file(s) if need be
	go mux.HandleFunc("/style.css", renderStaticFile("templates/style.css"))

	fmt.Println("now serving on http://0.0.0.0:5050")
	http.ListenAndServe(":5050", mux)
}
