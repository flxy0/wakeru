package hashgen

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"html/template"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/flxy0/wakeru/config"
	"github.com/flxy0/wakeru/helpers"
)

// This function takes care of generating new hashes, creating the directory for it,
// updating the directory list global variable and sending the hash to the user
func Generate(w http.ResponseWriter, r *http.Request) {
	fmt.Println(config.ConfigAllowGeneration)
	r.ParseMultipartForm(10)

	baseTmpl := template.Must(template.ParseFS(helpers.TemplateDir, "templates/base.gohtml"))
	genTmpl, genTmplErr := template.Must(baseTmpl.Clone()).ParseFS(helpers.TemplateDir, "templates/gen.gohtml")
	helpers.LogErr(genTmplErr)

	if r.FormValue("initGen") == "" {
		data := struct {
			AllowGenPage bool
			Error        string
			Hash         string
			InstanceName string
		}{
			AllowGenPage: config.ConfigAllowGeneration,
			Error:        "",
			Hash:         "",
			InstanceName: config.ConfigInstanceName,
		}

		tmplErr := genTmpl.Execute(w, data)
		helpers.LogErr(tmplErr)
	} else {
		currentTime := time.Now().Unix()

		rng := rand.Reader
		randInt, err := rand.Int(rng, big.NewInt(100000))
		helpers.LogErr(err)

		modifiedUnixTime := currentTime + randInt.Int64()

		shaHash := sha256.New()
		io.WriteString(shaHash, strconv.FormatInt(modifiedUnixTime, 10))

		// // To actually make use of the hash we need to format it into a string of hex digits
		hashString := fmt.Sprintf("%x", shaHash.Sum(nil))

		err = os.Mkdir(fmt.Sprintf("uploads/%s", hashString), 0777)
		helpers.LogErr(err)

		helpers.ServeDirs = helpers.FetchDirList()

		data := struct {
			AllowGenPage bool
			Error        string
			Hash         string
			InstanceName string
		}{
			AllowGenPage: config.ConfigAllowGeneration,
			Error:        "",
			Hash:         hashString,
			InstanceName: config.ConfigInstanceName,
		}

		tmplErr := genTmpl.Execute(w, data)
		helpers.LogErr(tmplErr)
	}
}
