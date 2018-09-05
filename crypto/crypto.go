package crypto

import (
	"encoding/base64"
	"github.com/onlyangel/apihandlers"
)

func Base64_encode(toEncode []byte) string {
	return base64.StdEncoding.EncodeToString(toEncode)
}

func Base64_decode(toDecode string) []byte {
	decoded, err := base64.StdEncoding.DecodeString(toDecode)
	apihandlers.PanicIfNil(err)
	return decoded
}
