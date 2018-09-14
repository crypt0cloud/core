package core

import (
	"fmt"
	_ "github.com/crypt0cloud/core/coordinator"
	"github.com/crypt0cloud/core/crypto"
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

	http.HandleFunc("/api/v1/pair_verification", apihandlers.RecoverApi(pair_verification))

	setup_hanlers()
	identification_handlers()
	transactions_handle()
	contracts_handler()
}

func warmup(w http.ResponseWriter, r *http.Request) {
	Warmup(r)
}
func Warmup(r *http.Request) {
	db := model.Open(r, "")

	if !db.IsRegisteredNodeID() {
		pubK, privK := ed_25519.GetNewKeyPair()
		node := new(md.NodeIdentification)
		node.Creation = time.Now().UnixNano()
		node.PublicKey = pubK
		node.PrivateKey = privK
		node.Myself = true

		db.RegisteredNodeID(node)
	}
}

func pong(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func pair_verification(w http.ResponseWriter, r *http.Request) {
	t := crypto.Validate_criptoTransaction(r.Body)
	if t == nil {
		fmt.Fprintf(w, "false")
	} else {
		fmt.Fprintf(w, "true")
	}
}
