package files

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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
	Error    string
	Feedback string
	Files    []UploadedFile
	UserHash string
}

// This function takes care of the view_files view.
// This view is for seeing which files have been uploaded using specific full(!) hash and deleting them.
// Initiating deletion flow depends on whether form values exist or not.
func ViewFiles(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	userHash := urlParts[2]

	r.ParseMultipartForm(10)

	var errList []string

	formError := r.ParseForm()
	if formError != nil {
		log.Println(formError)
	}

	formHash := r.FormValue("userHash")
	deleteVal := r.FormValue("deletion")

	if deleteVal != "" {
		deleteFiles(w, r, userHash)
		return
	}

	// ----------
	if formHash != "" && len(userHash) < 20 {
		http.Redirect(w, r, fmt.Sprintf("/viewfiles/%s", formHash), http.StatusSeeOther)
	} else if len(userHash) == 0 || userHash == "" {
		tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/viewfiles.gohtml"))
		tmpl.Execute(w, nil)
	} else if len(userHash) > 20 {
		tmplData, errString := computeFileListTmplData(userHash)

		if errString != "" {
			errList = append(errList, errString)
			tmplData.Error = strings.Join(errList, " & ")
		}

		tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/filelist.gohtml"))
		tmpl.Execute(w, tmplData)
	} else {
		fmt.Fprintf(w, "uh oh! something went wrong somewhere >_<")
	}
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
	if hash == "" {
		errList = append(errList, "ERROR: no hash specified")
	}

	// On disk path to the folder that's being uploaded too
	folderPath := "uploads/" + hash

	// Checks existing directories to see if the directory corresponding to the hash is present
	_, err := ioutil.ReadDir(folderPath)
	if err != nil {
		log.Println(err)
		errList = append(errList, "ERROR: no hash like that exists")
	}

	// Reads the file POSTed through the form so we can write it an actual file on disk.
	file, header, err := r.FormFile("upFile")
	if err != nil {
		log.Println(err)
		errList = append(errList, "ERROR: no file uploaded")
	}
	defer file.Close()

	// Creates a (not quite) temporary file to parse the bytes of the uploaded file into and store it on the disk.
	nameWithTimestamp := fmt.Sprintf("%d-%s", time.Now().Unix(), header.Filename)

	onDiskFile, err := os.OpenFile(fmt.Sprintf("%s/%s", folderPath, nameWithTimestamp), os.O_WRONLY|os.O_CREATE, 0666)
	defer onDiskFile.Close()
	io.Copy(onDiskFile, file)

	if err != nil {
		log.Println(err)
		errList = append(errList, "ERROR: couldn't write file")
		os.Remove(onDiskFile.Name())
	}

	// Print uploaded file info
	fmt.Printf("Uploaded file: %+v\n", nameWithTimestamp)

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
		filePathParts := strings.Split(onDiskFile.Name(), "/")
		http.Redirect(w, r, fmt.Sprintf("/uploads/%s/%s", filePathParts[1][:20], filePathParts[2]), http.StatusSeeOther)
	}
}

func deleteFiles(w http.ResponseWriter, r *http.Request, userHash string) {
	// Deletion flow!
	// Checks if the request includes the deletion Value and if so, deletes selected files in form.
	var errList []string
	var deleteFiles []string

	for k := range r.Form {
		if k != "deletion" {
			deleteFiles = append(deleteFiles, k)
		}
	}

	if len(deleteFiles) > 0 {
		for _, v := range deleteFiles {
			os.Remove(fmt.Sprintf("uploads/%s/%s", userHash, v))
			fmt.Println("deleted file " + fmt.Sprintf("uploads/%s/%s", userHash, v))
		}
	} else {
		errList = append(errList, "ERROR: no files were selected for deletion!")
	}

	tmplData, errString := computeFileListTmplData(userHash)
	tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/filelist.gohtml"))

	if errString != "" {
		errList = append(errList, errString)
		fmt.Println(errList)
		tmplData.Error = strings.Join(errList, " & ")
		tmpl.Execute(w, tmplData)
	} else if len(errList) != 0 {
		tmplData.Error = strings.Join(errList, " & ")
		tmpl.Execute(w, tmplData)
	} else {
		tmplData.Feedback = "selected files successfully deleted"
		tmpl.Execute(w, tmplData)
	}
	return
}

func computeFileListTmplData(userHash string) (fileListData FileListData, errStr string) {
	// Read the directory and return a list of the files in it, along with an error string if there is an error.
	files, err := ioutil.ReadDir(fmt.Sprintf("uploads/%s/", userHash))
	errString := ""

	if err != nil {
		log.Println(err)
		errString = "ERROR: there was an error retrieving files! are you sure you got the right hash?"
	}

	fileList := make([]UploadedFile, len(files))
	if errString == "" {
		if len(files) >= 1 {
			for i, v := range files {
				fileList[i] = UploadedFile{
					Name: v.Name(),
					URL:  fmt.Sprintf("/uploads/%s/%s", userHash[:20], v.Name()),
				}
			}
		} else {
			errString = "ERROR: currently no files uploaded"
		}
	}

	return FileListData{
		UserHash: userHash,
		Files:    fileList,
	}, errStr
}
