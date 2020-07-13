package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/onlyangel/apihandlers"
	"google.golang.org/appengine/log"

	crypto "github.com/crypt0cloud/crypto_go"

	"github.com/crypt0cloud/core/tools"

	md "github.com/crypt0cloud/model_go"

	"github.com/skip2/go-qrcode"
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

	ctx := tools.Context(r)
	log.Infof(ctx, "%+v", t)
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

	if !db.BlockExists(t.BlockSign) { // todo IS A BLOCK VALID ( THE LAST ONE )
		apihandlers.PanicWithMsg("Error in BlockSign")
	}

	if !db.AppIdExists(r, t.AppID) {
		apihandlers.PanicWithMsg("AppID doesnt exists")
	}

	// removing becaus this is signed by an app
	/*if user := db.UserSignExist(r, t.Signer); user == nil {
		apihandlers.PanicWithMsg("Sign doesnt exists")
	}*/

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

	qr := r.FormValue("qr")

	db := model.Open(r, "")

	id, err := strconv.ParseInt(idstr, 10, 64)
	apihandlers.PanicIfNotNil(err)

	t := db.GetSignRequest(r, id)

	jsonstr, err := json.Marshal(t)
	apihandlers.PanicIfNotNil(err)

	if qr == "true" {
		png, err := qrcode.Encode(string(jsonstr), qrcode.Medium, 256)
		apihandlers.PanicIfNotNil(err)
		w.Header().Add("Content-Type", "image/png")
		w.Write(png)
		return
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
	if t.Parent == "" {
		apihandlers.PanicWithMsg("Parent shouldnt be empty string")
	}
	if !db.AppIdExists(r, t.AppID) {
		apihandlers.PanicWithMsg("AppID doesnt exists")
	}

	parent := db.GetParentTransaction(r, t.Parent)
	if t.AppID != parent.AppID || t.Callback != parent.Callback || !tools.IsStringArrayEqual(t.SignerKinds, parent.SignerKinds) {
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
