package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/onlyangel/apihandlers"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/crypt0cloud/core/tools"
	crypto "github.com/crypt0cloud/crypto_go"
)

func setup_hanlers() {
	http.HandleFunc("/api/setup/configure_endpoint", apihandlers.RecoverApi(setup_configureEndPointName))
	http.HandleFunc("/api/setup/set_initial_node_registration", apihandlers.RecoverApi(setup_setInitialNodeRegistration))
}

func setup_configureEndPointName(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	endpoint := r.FormValue("endpoint")
	if db.SetupSetEndPointIfNull(endpoint) {
		fmt.Fprintf(w, "OK")
	} else {
		apihandlers.PanicWithMsg("Cant configure node")
	}
}

func setup_setInitialNodeRegistration(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	t := crypto.Validate_criptoTransaction(r.Body)

	log.Infof(appengine.NewContext(r), "NODE REGISTRATION: %+v", t)

	if t.SignKind != "__REGISTERNODE" {
		apihandlers.PanicWithMsg("Not Valid")
	}

	log.Infof(tools.Context(r), "SINGLE TRANSACTION")
	log.Infof(tools.Context(r), "%+v", t)

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

	myself := db.GetNodeId()

	if t.ToNode.PublicKey != myself.PublicKey {
		apihandlers.PanicWithMsg("Different Origin Node")
	}

	db.SetupNodeRegistrationInDeployment(t)

	jsonstr, err := json.Marshal(t)
	apihandlers.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", jsonstr)
}
