package files

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// -----------------------------------------------------------------------------------------
// This code handles all the file related actions.
// Such as viewing which files were uploaded by a given hash, and deleting them if need be.
// -----------------------------------------------------------------------------------------

type UploadedFile struct {
	Name string
	URL  string
}

// File List Template Data struct
type FileListData struct {
	UserHash string
	Files    []UploadedFile
	Error    string
}

// This function takes care of the view_files view.
// This view is for seeing which files have been uploaded using specific full(!) hash and deleting them.
func ViewFiles(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	userHash := urlParts[2]

	errString := ""

	formError := r.ParseForm()
	if formError != nil {
		log.Println(formError)
	}

	formHash := r.FormValue("userHash")

	if formHash != "" {
		http.Redirect(w, r, fmt.Sprintf("/viewfiles/%s", formHash), http.StatusSeeOther)
	} else if len(userHash) == 0 || userHash == "" {
		tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/viewfiles.gohtml"))

		tmplErr := tmpl.Execute(w, nil)
		if tmplErr != nil {
			log.Println(tmplErr)
		}
	} else if len(userHash) > 20 {
		files, err := ioutil.ReadDir(fmt.Sprintf("uploads/%s/", userHash))

		if err != nil {
			log.Println(err)
			errString = "uh oh... there was an error retrieving files!"
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

		if errString != "" {
			tmplData.Error = errString
		}

		tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/filelist.gohtml"))

		tmplErr := tmpl.Execute(w, tmplData)
		if tmplErr != nil {
			log.Println(tmplErr)
		}
	} else {
		fmt.Fprintf(w, "uh oh! something went wrong somewhere >_<")
	}
}

func DeleteFiles(w http.ResponseWriter, r *http.Request) {
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

// This function returns either the standard upload template, or a redirect to the File after it has been uploaded.
// It handles both of those cases to not use another url path and to make having errors render after issues easier.
func Upload(w http.ResponseWriter, r *http.Request) {
	// Slice, even though name is List, for appending errors to.
	var errList []string

	// We need to parse the multi part form before we can even evaluate what "kind of request" it is
	r.ParseMultipartForm(10 << 20)

	// If the page isn't loaded with form data from the upload, render regular template.
	if len(r.PostForm) == 0 {
		tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/upload.gohtml"))

		tmplErr := tmpl.Execute(w, nil)
		if tmplErr != nil {
			log.Println(tmplErr)
			return
		}
		return
	}

	// verifies a hash is present
	hash := r.FormValue("userHash")
	fmt.Println(hash)
	if hash == "" {
		errList = append(errList, "ERROR: no hash specified")
	}

	// Checks existing directories to see if the directory corresponding to the hash is present
	_, err := ioutil.ReadDir("uploads/" + hash)
	if err != nil {
		fmt.Println("something wrong with the upload folder/hash")
		log.Println(err)
		errList = append(errList, "ERROR: no hash like that exists")
	}

	// On disk path to the folder that's being uploaded too
	folderPath := "uploads/" + hash

	// Reads the file POSTed through the form so we can write it an actual file on disk.
	file, handler, err := r.FormFile("upFile")
	if err != nil {
		log.Println(err)
		errList = append(errList, "ERROR: no file uploaded")
	}
	defer file.Close()

	// Print uploaded file info
	fmt.Printf("Uploaded file: %+v\n", handler.Filename)

	// Creates a (not quite) temporary file to parse the bytes of the uploaded file into and store it on the disk.
	tempFile, err := ioutil.TempFile(folderPath, "*-"+handler.Filename)
	fmt.Println(tempFile.Name())
	if err != nil {
		log.Println(err)
		errList = append(errList, "ERROR: writing file")
		defer tempFile.Close()
		os.Remove(tempFile.Name())
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
	}

	tempFile.Write(fileBytes)
	defer tempFile.Close()

	// Returns template with error message if something went wrong
	if len(errList) > 0 {
		tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/upload.gohtml"))

		tmplData := struct {
			Error string
		}{
			Error: strings.Join(errList, " & "),
		}

		tmplErr := tmpl.Execute(w, tmplData)
		if tmplErr != nil {
			log.Println(tmplErr)
		}

	} else {
		filePathParts := strings.Split(tempFile.Name(), "/")
		http.Redirect(w, r, fmt.Sprintf("/uploads/%s/%s", filePathParts[1][:20], filePathParts[2]), http.StatusSeeOther)
	}
}

// This function handles the post form from the upload.gohtml
// It verifies that there's a valid hash and a file present upon submission
// If not, it will simply display a simple text message informing the user what is wrong
func UploadResponse(w http.ResponseWriter, r *http.Request) {
}
