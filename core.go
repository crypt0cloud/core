package core

import (
	"fmt"

	md "github.com/crypt0cloud/core/model"
	_ "github.com/crypt0cloud/model_datastore"

	_ "github.com/crypt0cloud/core/coordinator"

	"github.com/onlyangel/apihandlers"
	"log"
	"net/http"
)

var model md.ModelConnector

func init(){
	var err error
	model, err = md.Open("datastore")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/ping", apihandlers.RecoverApi(pong))
}

func pong(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "pong")
}

