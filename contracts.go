package core

import (
	"encoding/json"
	"fmt"
	"github.com/crypt0cloud/core/crypto"
	"github.com/crypt0cloud/core/tools"
	"github.com/onlyangel/apihandlers"
	"net/http"
)

func contracts_handler() {
	http.HandleFunc("/api/v1/create_contract", apihandlers.RecoverApi(contracts_createContract))
	http.HandleFunc("/api/v1/sign_contract", apihandlers.RecoverApi(contracts_sign))

	http.HandleFunc("/api/v1/create_signingRequest", apihandlers.RecoverApi(contract_createSigningRequest))
	http.HandleFunc("/api/v1/get_signingRequest", apihandlers.RecoverApi(contract_getSigningRequest))

}

func contracts_createContract(w http.ResponseWriter, r *http.Request) {
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

func contracts_sign(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	t := crypto.Validate_criptoTransaction(r.Body)

	signreq := db.GetSignRequest(r, t.IdVal)

	if t.Parent != signreq.Parent || t.ParentBlock != signreq.ParentBlock || t.SignKind != signreq.SignKind || t.AppID != signreq.AppID || t.Payload != signreq.Payload || t.Callback != signreq.Callback || !tools.IsStringArrayEqual(t.SignerKinds, signreq.SignerKinds) {
		panic(fmt.Errorf("Values of the signing request are different from the parent transaction"))
	}

	t = db.InsertTransaction(r, t)
	jsonstr, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "%s", jsonstr)

	//TODO SEND CALLBACK
}
