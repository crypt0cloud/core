package query

import (
	"net/http"

	"github.com/onlyangel/apihandlers"
)

func _get_authentication(r *http.Request) *authentication {
	apppublickey := r.Header.Get("C0C_Auth")
	if apppublickey == "" {
		apihandlers.PanicWithMsg("Invalid Credentials")
	}

	return NewAuthentiation(apppublickey)
}

type authentication struct {
	AppPublicKey string
}

func NewAuthentiation(apk string) (auth *authentication) {
	auth.AppPublicKey = apk
	return

}
