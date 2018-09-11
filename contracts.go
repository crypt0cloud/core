package core

import (
	"github.com/onlyangel/apihandlers"
	"net/http"
)

func contracts_handler() {
	//http.HandleFunc("/api/v1/create_contract",apihandlers.Recover(contracts_createContract))
	//http.HandleFunc("/api/v1/sign_contract",apihandlers.Recover(contracts_sign))

	http.HandleFunc("/api/v1/create_signingRequest", apihandlers.RecoverApi(contract_createSigningRequest))
	http.HandleFunc("/api/v1/get_signingRequest", apihandlers.RecoverApi(contract_getSigningRequest))

}
