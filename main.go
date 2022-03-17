package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"sr.ht/flxy/wakeru/hashgen"
	"sr.ht/flxy/wakeru/helpers"
)

type UploadedFile struct {
	Name string
	URL  string
}

// File List Template Data struct
type FileListData struct {
	UserHash string
	Files    []UploadedFile
}

// Type for Template Files
type templateFile struct {
	name     string
	contents string
}

// Global variables for the file view templates so they remain cached but still accessible in the function without explicitly passing
var view_files_tmpl = template.Must(template.ParseFiles("templates/view_files.gohtml"))
var file_list_tmpl = template.Must(template.ParseFiles("templates/file_list.gohtml"))

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
func uploaded(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	// verifies a hash is present
	hash := r.FormValue("userHash")
	fmt.Println(hash)
	if hash == "" {
		fmt.Fprintf(w, "ERROR: no hash specified")
		return
	}

	// Checks existing directories to see if the directory corresponding to the hash is present
	_, err := ioutil.ReadDir("uploads/" + hash)
	if err != nil {
		fmt.Println("something wrong with the upload folder/hash")
		fmt.Println(err)
		fmt.Fprintf(w, "ERROR: no hash like that exists")
		return
	}

	folderPath := "uploads/" + hash

	file, handler, err := r.FormFile("upFile")
	if err != nil {
		fmt.Println("error receiving file")
		fmt.Println(err)
		fmt.Fprintf(w, "ERROR: no file uploaded")
		return
	}
	defer file.Close()

	// Print uploaded file info
	fmt.Printf("Uploaded file: %+v\n", handler.Filename)

	tempFile, err := ioutil.TempFile(folderPath, "*-"+handler.Filename)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "ERROR: writing file")
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	tempFile.Write(fileBytes)

	fmt.Fprintf(w, "succesfully wrote new file")
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
		if strings.HasPrefix(v, userHash) && len(userHash) == 20 {
			dirMatch = v
		} else {
			fmt.Fprintf(w, "Hash is wrong or doesn't exist")
			return
		}
	}

	filePath := fmt.Sprintf("uploads/%s/%s", dirMatch, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(w, "File does not exist")
		return
	}
	http.ServeFile(w, r, filePath)
}

// This function takes care of the view_files view.
// This view is for seeing which files have been uploaded by a specific hash and deleting them.
func viewFiles(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	userHash := urlParts[2]

	error := r.ParseForm()
	if error != nil {
		fmt.Println(error)
	}
	fmt.Println(r.FormValue("deletion"))

	if r.FormValue("userHash") != "" {
		hash := r.FormValue("userHash")
		http.Redirect(w, r, fmt.Sprintf("/view_files/%s", hash), http.StatusSeeOther)
	} else if len(userHash) == 0 {
		err := view_files_tmpl.Execute(w, nil)

		if err != nil {
			fmt.Println("Error rendering a template")
		}
	} else if len(userHash) > 20 {
		files, err := ioutil.ReadDir(fmt.Sprintf("uploads/%s/", userHash))

		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Error retreiving files")
			return
		}

		fileList := make([]UploadedFile, len(files))
		for i, v := range files {
			fileList[i] = UploadedFile{
				Name: v.Name(),
				URL:  fmt.Sprintf("/uploads/%s/%s", userHash[:20], v.Name()),
			}
		}

		tmplData := FileListData{
			UserHash: userHash,
			Files:    fileList,
		}
		err = file_list_tmpl.Execute(w, tmplData)
		if err != nil {
			fmt.Println("Error rendering template")
			fmt.Fprintf(w, "Error providing the page!")
		}
	} else {
		fmt.Println("something went wrong oops")
		return
	}
}

func deleteFiles(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Error receiving data")
		return
	}

	refererSlice := strings.Split(r.Referer(), "/")
	userHash := refererSlice[len(refererSlice)-1]

	bodyStr := string(body)
	formFiles := strings.Split(bodyStr, "&")
	fileNames := make([]string, len(formFiles))

	for i, v := range formFiles {
		if len(v) > 0 {
			fileNames[i] = strings.Split(v, "=")[0]
		}
	}

	for _, v := range fileNames {
		err := os.Remove(fmt.Sprintf("uploads/%s/%s", userHash, v))
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Error deleting file(s)")
			return
		}
	}

	fmt.Fprintf(w, "Files succesfully deleted")
}

func main() {
	var (
		// template_list    = []string{"templates/base.gothml", "templates/index.gohtml"}
		index_tmpl = template.Must(template.ParseFiles("templates/base.gohtml", "templates/index.gohtml"))
		// templates =      template.Must(template.ParseFiles("templates/base.gohtml"))
		// index_tmpl, _ = template.Must(base_tmpl.Clone()).ParseFiles("templates/index.gohtml")
		gen_tmpl = template.Must(template.ParseFiles("templates/base.gohtml", "templates/gen.gohtml"))
	)

	// template.Must()

	mux := http.NewServeMux()

	// index route
	go mux.HandleFunc("/", renderDatalessTemplate(*index_tmpl))

	// // hash generation related routes
	go mux.HandleFunc("/generate", renderDatalessTemplate(*gen_tmpl))
	go mux.HandleFunc("/generated", hashgen.Generated)

	// // upload form and handling of post errors
	// go mux.HandleFunc("/upload", renderDatalessTemplate(*upload_tmpl))
	// go mux.HandleFunc("/uploaded", uploaded)

	// // serve uploads
	// go mux.HandleFunc("/uploads/", handleServeContent)

	// // view files corresponding to hash
	// go mux.HandleFunc("/view_files/", viewFiles)
	// go mux.HandleFunc("/view_files/deletion", deleteFiles)
	go mux.HandleFunc("/style.css", renderStaticFile("templates/style.css"))

	fmt.Println("now serving on http://0.0.0.0:5050")
	http.ListenAndServe(":5050", mux)
}
