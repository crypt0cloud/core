package coordinator

import (
	"encoding/base64"
	"github.com/crypt0cloud/core/tools"
	"github.com/onlyangel/apihandlers"
	"golang.org/x/crypto/ed25519"
	"log"
	"math/rand"
	"net/http"
	"time"

	md "github.com/crypt0cloud/core/model"
)

var model md.ModelConnector

func init(){
	var err error
	model, err = md.Open("datastore")
	if err != nil {
		log.Fatal(err)
	}
	//http.HandleFunc("/api/api",api_handler)
	http.HandleFunc("/api/coord/register_masterkey",apihandlers.RecoverApi(coord_registerMasterKey))
}

func coord_registerMasterKey(w http.ResponseWriter, r *http.Request){
	//ctx := tools.Context(r)
	db := model.Open(r,"")
	key_exists := db.Coord_MasterKeyExists()

	if key_exists {
		apihandlers.PanicWithMsg("Already asigned key")
	}

	key := tools.FormValueEscaped(r,"key")
	url := tools.FormValueEscaped(r,"url")
	if key == "" || url == "" {
		apihandlers.PanicWithMsg("In parameters")
	}

	keyarr, err := base64.StdEncoding.DecodeString(key)
	apihandlers.PanicIfNil(err)

	makey := new(md.MasterKey)
	makey.MasterPublicKey = keyarr
	makey.URL = url
	makey.CoordinatorPublic, makey.CoordinatorPrivate, err = ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))

	db.Coord_InsertKey(makey)

	tools.PrintJson(w, makey)
}