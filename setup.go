package core

import (
	"fmt"
	"github.com/onlyangel/apihandlers"
	"net/http"
)

func setup_hanlers() {
	http.HandleFunc("/api/setup/configure_endpoint", apihandlers.RecoverApi(setup_configureEndPointName))
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
