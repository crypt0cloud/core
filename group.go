package core

import (
	"encoding/json"
	"fmt"
	"github.com/crypt0cloud/core/crypto"
	"github.com/crypt0cloud/core/tools"
	"github.com/onlyangel/apihandlers"
	"net/http"
	"strconv"
	"time"

	md "github.com/crypt0cloud/core/model"
)

func group_handler() {
	http.HandleFunc("/api/v1/create_group", apihandlers.RecoverApi(group_createGroup))

	http.HandleFunc("/api/v1/create_signingRequest", apihandlers.RecoverApi(group_createSigningRequest))
	http.HandleFunc("/api/v1/get_signingRequest", apihandlers.RecoverApi(group_getSigningRequest))
	http.HandleFunc("/api/v1/sign_signingRequest", apihandlers.RecoverApi(group_sign_signingRequest))

}

func group_createGroup(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	t := crypto.Validate_criptoTransaction(r.Body)

	//TODO: verify the the content and the values coincide

	if t.Parent != "" { //request should be 0
		apihandlers.PanicWithMsg("New Contract should have a parent = 0")
	}

	if t.SignKind != "__NEWCONTRACT" {
		apihandlers.PanicWithMsg("Just new Contracts are allowed")
	}

	signerkidnslen := len(t.SignerKinds)
	if signerkidnslen == 0 {
		apihandlers.PanicWithMsg("SignerKinds had to have at least one SignerKind")
	}

	if !db.AppIdExists(r, t.AppID) {
		apihandlers.PanicWithMsg("AppID doesnt exists")
	}

	if user := db.UserSignExist(r, t.Signer); user == nil {
		apihandlers.PanicWithMsg("Sign doesnt exists")
	}

	myself := db.GetNodeId()
	if t.FromNode.PublicKey != myself.PublicKey {
		apihandlers.PanicWithMsg("Different Origin Node")
	}

	t = db.InsertTransaction(r, t)
	jsonstr, err := json.Marshal(t)
	apihandlers.PanicIfNotNil(err)
	fmt.Fprintf(w, "%s", string(jsonstr))

	//TODO SEND CALLBACK
}

func group_sign_signingRequest(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	t := crypto.Validate_criptoTransaction(r.Body)

	signreq := db.GetSignRequest(r, t.IdVal)

	if t.Parent != signreq.Parent || t.SignKind != signreq.SignKind || t.AppID != signreq.AppID || t.Payload != signreq.Payload || t.Callback != signreq.Callback || !tools.IsStringArrayEqual(t.SignerKinds, signreq.SignerKinds) {
		apihandlers.PanicWithMsg("Values of the signing request are different from the parent transaction")
	}

	t = db.InsertTransaction(r, t)
	jsonstr, err := json.Marshal(t)
	apihandlers.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", jsonstr)

	//TODO SEND CALLBACK
}

func group_getSigningRequest(w http.ResponseWriter, r *http.Request) {
	idstr := r.FormValue("id")
	if idstr == "" {
		apihandlers.PanicWithMsg("Problem with parameters")
	}

	db := model.Open(r, "")

	id, err := strconv.ParseInt(idstr, 10, 64)
	apihandlers.PanicIfNotNil(err)

	t := db.GetSignRequest(r, id)

	jsonstr, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "%s", string(jsonstr))
}

func group_createSigningRequest(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	bodydecoder := json.NewDecoder(r.Body)

	t := new(md.Transaction)
	err := bodydecoder.Decode(t)
	defer r.Body.Close()

	if err != nil {
		if err.Error() == "EOF" {
			apihandlers.PanicWithMsg("Empty body")
		} else {
			panic(err)
		}
	}

	if t.FromNode.PublicKey == "" {
		apihandlers.PanicWithMsg("From node shouldnt be empty")
	}
	if t.Content != "" {
		apihandlers.PanicWithMsg("Content shoudl be empty string")
	}
	if t.Hash != "" {
		apihandlers.PanicWithMsg("Hash shoudl be empty string")
	}
	if t.Signer != "" {
		apihandlers.PanicWithMsg("Signer shoudl be empty string")
	}
	if t.Sign != "" {
		apihandlers.PanicWithMsg("Sign shoudl be empty string")
	}
	if t.InsertMoment != 0 {
		apihandlers.PanicWithMsg("InsertMoment shoudl be 0")
	}
	if t.Parent == "" {
		apihandlers.PanicWithMsg("Parent shoudl be empty string")
	}
	if !db.AppIdExists(r, t.AppID) {
		apihandlers.PanicWithMsg("AppID doesnt exists")
	}

	parent := db.GetParentTransaction(r, t.Parent)
	if t.AppID != parent.AppID || t.Payload != parent.Payload || t.Callback != parent.Callback || !tools.IsStringArrayEqual(t.SignerKinds, parent.SignerKinds) {
		apihandlers.PanicWithMsg("Values of the signing request are different from the parent transaction")
	}

	if !tools.IsStringInArray(t.SignKind, t.SignerKinds) {
		apihandlers.PanicWithMsg("SignKind is not in SignerKinds")
	}

	t.Creation = time.Now().UnixNano()

	t = db.InsertSignRequest(r, t)
	jsonstr, err := json.Marshal(t)
	apihandlers.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", string(jsonstr))
}
