package hashgen

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"sr.ht/flxy/wakeru/helpers"
)

// This function takes care of generating new hashes, creating the directory for it,
// updating the directory list global variable and sending the hash to the user
func Generated(w http.ResponseWriter, r *http.Request) {
	// exitWithError := func(err error) http.HandlerFunc {
	// 	return func(w http.ResponseWriter, r *http.Request) {
	// 		log.Fatal(err)
	// 		fmt.Fprint(w, "error generating hash! sorry!")
	// 	}
	// }
	exitWithError := func(err error) {
		log.Fatalln(err)
	}

	currentTime := time.Now().Unix()

	rng := rand.Reader
	randInt, err := rand.Int(rng, big.NewInt(100000))
	if err != nil {
		exitWithError(err)
	}

	modifiedUnixTime := currentTime + randInt.Int64()

	shaHash := sha256.New()
	io.WriteString(shaHash, strconv.FormatInt(modifiedUnixTime, 10))

	// To actually make use of the hash we need to format it into a string of hex digits
	hashString := fmt.Sprintf("%x", shaHash.Sum(nil))

	err = os.Mkdir(fmt.Sprintf("uploads/%s", hashString), 0777)
	if err != nil {
		exitWithError(err)
	}

	helpers.ServeDirs = helpers.FetchDirList()

	data := struct {
		Hash string
	}{
		Hash: hashString,
	}

	tmpl := template.Must(template.ParseFiles("templates/base.gohtml", "templates/generated.gohtml"))

	tmpl_err := tmpl.Execute(w, data)
	if tmpl_err != nil {
		log.Fatal(err)
	}
}
