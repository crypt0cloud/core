package core

import (
	"encoding/json"
	"fmt"
	md "github.com/crypt0cloud/core/model"
	"github.com/crypt0cloud/core/tools"
	"github.com/onlyangel/apihandlers"
	"net/http"
	"strconv"
	"time"
)

func contract_getSigningRequest(w http.ResponseWriter, r *http.Request) {
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

func contract_createSigningRequest(w http.ResponseWriter, r *http.Request) {
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
