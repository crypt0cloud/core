package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/onlyangel/apihandlers"
	"google.golang.org/appengine/log"

	"github.com/crypt0cloud/core/connections"
	"github.com/crypt0cloud/core/crypto"
	"github.com/crypt0cloud/core/tools"
	md "github.com/crypt0cloud/model_go"
)

func transactions_handle() {

	http.HandleFunc("/api/v1/post_single_transaction", apihandlers.RecoverApi(transactions_postSingleTransaction))

}

func transactions_postSingleTransaction(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	t := crypto.Validate_criptoTransaction(r.Body)

	log.Infof(tools.Context(r), "SINGLE TRANSACTION")
	log.Infof(tools.Context(r), "%+v", t)

	//TODO: verify the the content and the values coincide

	coord_key := new(md.Transaction)

	if t.Parent != "" {
		apihandlers.PanicWithMsg("Single transaction should have a parent = \"\"")
	}
	if t.Creation == 0 {
		apihandlers.PanicWithMsg("Creation should be a current time")
	}

	if t.BlockSign == "" {
		apihandlers.PanicWithMsg("Every transaction must have a BlockSign")
	}
	//TODO: verify time

	if t.SignKind == "__REGISTERNODE" {
		if t.Signer != t.AppID {
			apihandlers.PanicWithMsg("In sign credentials order")
		}
		if t.Payload == "" {
			apihandlers.PanicWithMsg("Error in sign credentials order")
		}
		if len(t.SignerKinds) != 1 {
			apihandlers.PanicWithMsg("Single transaction should have a parent = 0")
		}

		if t.SignKind != t.SignerKinds[0] {
			apihandlers.PanicWithMsg("Single transaction should have a parent = 0")
		}

		//todo verify the callback it is a url
		if t.Callback == "" {
			apihandlers.PanicWithMsg("There should be a callback")
		}

		if db.NodeExists(r) {
			apihandlers.PanicWithMsg("Node Already setted up")
		}

	} else {
		coord_key = db.GetCoordinatorKey(r)
	}

	if t.SignKind == "__NEWAPP" {
		if t.Payload != "__NEWAPP" {
			apihandlers.PanicWithMsg("Error in sign credentials order")
		}
		if len(t.SignerKinds) != 1 {
			apihandlers.PanicWithMsg("Single transaction should have a parent = 0")
		}

		if t.SignKind != t.SignerKinds[0] {
			apihandlers.PanicWithMsg("Single transaction should have a parent = 0")
		}

		//todo verify the callback it is a url
		if t.Callback == "" {
			apihandlers.PanicWithMsg("There should be a callback")
		}

		if db.AppIdExists(r, t.AppID) {
			apihandlers.PanicWithMsg("App sign already exists")
		}

		if t.Signer != coord_key.Signer {
			apihandlers.PanicWithMsg("Un authoriced request")
		}
	}

	if t.SignKind == "NewUser" {
		if t.Signer != t.AppID {
			apihandlers.PanicWithMsg("In sign credentials order")
		}
		if t.Payload == "" {
			apihandlers.PanicWithMsg("Error in sign credentials order")
		}
		if len(t.SignerKinds) != 1 {
			apihandlers.PanicWithMsg("Single transaction should have a parent = 0")
		}

		if t.SignKind != t.SignerKinds[0] {
			apihandlers.PanicWithMsg("Single transaction should have a parent = 0")
		}

		//todo verify the callback it is a url
		if t.Callback == "" {
			apihandlers.PanicWithMsg("There should be a callback")
		}

		if user := db.UserSignExist(r, t.AppID); user != nil {
			apihandlers.PanicWithMsg("Sign exists")
		}

		if user := db.UserPayloadExist(r, t.Payload); user != nil {
			apihandlers.PanicWithMsg("User Exists")
		}

	}
	myself := db.GetNodeId()

	if t.ToNode.PublicKey != myself.PublicKey {
		apihandlers.PanicWithMsg("Different Origin Node")
	}

	if t.SignKind != "__REGISTERNODE" {
		t.External = false
		log.Infof(tools.Context(r), " COORD KEY ")
		log.Infof(tools.Context(r), "%+v", coord_key)
		sino := connections.ValidateTransactionWithPeers(r, t, coord_key.FromNode)
		if !sino {
			apihandlers.PanicWithMsg("Peer virification invalid")
		}
	}

	t = db.InsertTransaction(r, t)

	jsonstr, err := json.Marshal(t)
	apihandlers.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", jsonstr)

	//TODO: enviar la transaccion exitosa al call back
}
