package query

import (
	"fmt"
	"github.com/onlyangel/apihandlers"
	"net/http"
)

func init() {
	http.HandleFunc("query/v1/user/exists", apihandlers.RecoverApi(query_user_exists))
}

func query_user_exists(w http.ResponseWriter, r *http.Request) {
	public_key := r.FormValue("public_key")

	if public_key == "" {
		apihandlers.PanicWithMsg("In parameters")
	}

	db := model.Open(r, "")

	valid := db.UserSignExist(r, public_key)

	if valid == nil {
		fmt.Fprint(w, "false")
	} else {
		fmt.Fprint(w, "true")
	}
}
