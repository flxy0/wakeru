package hashgen

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/flxy0/wakeru/helpers"
)

// This function takes care of generating new hashes, creating the directory for it,
// updating the directory list global variable and sending the hash to the user
func Generate(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10)

	if r.FormValue("initGen") == "" {
		tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/gen.gohtml"))

		tmplErr := tmpl.Execute(w, nil)
		if tmplErr != nil {
			log.Println(tmplErr)
		}
	} else {
		currentTime := time.Now().Unix()

		rng := rand.Reader
		randInt, err := rand.Int(rng, big.NewInt(100000))
		if err != nil {
			log.Println(err)
		}

		modifiedUnixTime := currentTime + randInt.Int64()

		shaHash := sha256.New()
		io.WriteString(shaHash, strconv.FormatInt(modifiedUnixTime, 10))

		// // To actually make use of the hash we need to format it into a string of hex digits
		hashString := fmt.Sprintf("%x", shaHash.Sum(nil))

		err = os.Mkdir(fmt.Sprintf("uploads/%s", hashString), 0777)
		if err != nil {
			log.Println(err)
		}

		helpers.ServeDirs = helpers.FetchDirList()

		data := struct {
			Error string
			Hash  string
		}{
			Error: "",
			Hash:  hashString,
		}

		tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/gen.gohtml"))

		tmplErr := tmpl.Execute(w, data)
		if tmplErr != nil {
			log.Println(tmplErr)
		}
	}
}
