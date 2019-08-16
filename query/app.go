package query

import "net/http"

func app_transactions(w http.ResponseWriter, r *http.Request){
	_get_authentication(r)
}
