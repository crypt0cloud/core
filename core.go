package core

import (
	"fmt"
	_ "github.com/crypt0cloud/core/coordinator"
	"github.com/crypt0cloud/core/crypto/ed_25519"
	md "github.com/crypt0cloud/core/model"
	"github.com/onlyangel/apihandlers"
	"log"
	"net/http"
	"time"
)

var model md.ModelConnector

func init() {
	var err error
	model, err = md.Open("datastore")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/ping", apihandlers.RecoverApi(pong))
	http.HandleFunc("/_ah/warmup", apihandlers.RecoverApi(warmup))
}

func warmup(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	if !db.IsRegisteredNodeID() {
		pubK, privK := ed_25519.GetNewKeyPair()
		node := new(md.NodeIdentification)
		node.Creation = time.Now().UnixNano()
		node.PublicKey = pubK
		node.PrivateKey = privK

		db.RegisteredNodeID(node)
	}

}

func pong(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}
