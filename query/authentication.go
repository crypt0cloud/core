package query

import (
	"github.com/onlyangel/apihandlers"
	"net/http"
)



func _get_authentication(r *http.Request) *authentication {
	apppublickey := r.Header.Get("Crypt0.Cloud-AppId")
	if apppublickey == ""{
		apihandlers.PanicWithMsg("Invalid Credentials")
	}

	return NewAuthentiation(apppublickey)
}

type authentication struct {
	AppPublicKey	string
}

func NewAuthentiation(apk string) ( auth *authentication) {
	auth.AppPublicKey = apk
	return

}