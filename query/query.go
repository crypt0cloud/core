package query

import (
	md "github.com/crypt0cloud/core/model"
	"github.com/onlyangel/apihandlers"
	"log"
	"net/http"
)

var model md.ModelConnector

func init() {
	var err error
	model, err = md.Open("datastore")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/query/v1/blocks", apihandlers.RecoverApi(blocks))
}

func blocks(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	size, offset := _handleFilters(r)

	blocks := db.GetBlocksByOffset(size, offset)

	apihandlers.WriteAsJsonList(w, blocks)
}
