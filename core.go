package core

import (
	"fmt"
	"github.com/crypt0cloud/core/coordinator"
	"log"
	"net/http"
	"time"

	"github.com/onlyangel/apihandlers"

	crypto "github.com/crypt0cloud/crypto_go"
	"github.com/crypt0cloud/crypto_go/ed_25519"
	md "github.com/crypt0cloud/model_go"

	_ "github.com/crypt0cloud/core/coordinator"
	_ "github.com/crypt0cloud/core/query"
)

var model md.ModelConnector

func init() {

	http.HandleFunc("/ping", apihandlers.RecoverApi(pong))
	http.HandleFunc("/_ah/warmup", apihandlers.RecoverApi(warmup))

	http.HandleFunc("/api/v1/pair_verification", apihandlers.RecoverApi(pair_verification))

	setup_hanlers()
	identification_handlers()
	transactions_handle()
	group_handler()

	block_handlers()

	query_handlers()
}

func InitDB() error {
	var err error
	var mo *md.ModelConnector
	mo, err = md.OpenDefault()
	model = *mo
	log.Printf("initdb: %+v", mo)
	coordinator.InitDB()
	return err
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
		db := model.Open(r, "")

		fmt.Fprintf(w, "true")

		t.External = true
		t = db.InsertPairVerificationTransaction(r, t)
	}
}
