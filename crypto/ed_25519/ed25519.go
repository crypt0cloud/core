package ed_25519

import (
	"github.com/crypt0cloud/core/crypto"
	"github.com/onlyangel/apihandlers"
	"golang.org/x/crypto/ed25519"
	"math/rand"
	"time"
)

func GetNewKeyPair() (string, string) {
	appPublicKey, appPrivateKey, err := ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))
	apihandlers.PanicIfNotNil(err)

	return crypto.Base64_encode(appPublicKey), crypto.Base64_encode(appPrivateKey)
}
