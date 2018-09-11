package core

import (
	"encoding/json"
	"fmt"
	"github.com/crypt0cloud/core/tools"
	"github.com/onlyangel/apihandlers"
	"google.golang.org/appengine/log"
	"net/http"
)

func identification_handlers() {
	//http.HandleFunc("/api/v1/last_block",apihandlers.Recover(identification_getNodeId))
	http.HandleFunc("/api/v1/node_id", apihandlers.Recover(identification_getNodeId))
}

func identification_getNodeId(w http.ResponseWriter, r *http.Request) {
	ctx := tools.Context(r)
	db := model.Open(r, "")

	block := db.GetNodeId()

	// CHANGE TO NODE IDENTIFICATION AND CHANGE MYSELF TO FALSE
	block.Myself = false

	log.Debugf(ctx, "%v", block)

	jsonstr, err := json.Marshal(block)
	apihandlers.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", string(jsonstr))
}
