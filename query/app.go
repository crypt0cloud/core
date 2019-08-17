package query

import (
	"fmt"
	"github.com/onlyangel/apihandlers"
	"net/http"
)

func init() {
	http.HandleFunc("/query/v1/app/transactions", apihandlers.RecoverApi(app_transactions))
	http.HandleFunc("/query/v1/app/transaction", apihandlers.RecoverApi(app_transaction))

}

func app_transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	publickey := r.FormValue("key")
	sign := r.FormValue("sign")
	clear := r.FormValue("clear")

	if publickey == "" || sign == "" || clear == "" {
		panic(fmt.Errorf("In Parameters"))
	}

	from := r.FormValue("from")
	to := r.FormValue("to")

	meta := r.FormValue("metadata")
	metadata := false
	if meta == "true" {
		metadata = true
	}

	db := model.Open(r, "")
	tr := db.GetApplicationTransactions(publickey, from, to, metadata)

	if len(tr) == 0 {
		fmt.Fprint(w, "[]")
		return
	}
	apihandlers.WriteAsJsonList(w, tr)
}

func app_transaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	publickey := r.FormValue("key")
	sign := r.FormValue("sign")
	clear := r.FormValue("clear")

	trsign := r.FormValue("trsign")

	if publickey == "" || sign == "" || clear == "" || trsign == "" {
		panic(fmt.Errorf("In Parameters"))
	}

	meta := r.FormValue("metadata")
	metadata := false
	if meta == "true" {
		metadata = true
	}

	db := model.Open(r, "")
	tr := db.GetApplicationTransaction(publickey, trsign, metadata)
	if tr == nil {
		panic(fmt.Errorf("Not valid Sign"))
	}

	apihandlers.WriteAsJson(w, tr)
}
