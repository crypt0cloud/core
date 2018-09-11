package crypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/crypt0cloud/core/model"
	"github.com/onlyangel/apihandlers"
	"golang.org/x/crypto/ed25519"
	"io"
)

func Base64_encode(toEncode []byte) string {
	return base64.StdEncoding.EncodeToString(toEncode)
}

func Base64_decode(toDecode string) []byte {
	decoded, err := base64.StdEncoding.DecodeString(toDecode)
	apihandlers.PanicIfNotNil(err)
	return decoded
}

func Validate_criptoTransaction(readbody io.ReadCloser) *model.Transaction {
	bodydecoder := json.NewDecoder(readbody)

	t := new(model.Transaction)
	err := bodydecoder.Decode(t)
	defer readbody.Close()

	if err != nil {
		if err.Error() == "EOF" {
			panic("Empty body")
		} else {
			panic(err)
		}
	}

	//t.InsertMoment = time.Now().UnixNano()

	content := Base64_decode(t.Content)

	hash := Base64_decode(t.Hash)

	sign := Base64_decode(t.Sign)

	signer := Base64_decode(t.Signer)

	sha_256 := sha256.New()
	sha_256.Write([]byte(content))
	contentsha := sha_256.Sum(nil)
	if !bytes.Equal(contentsha, hash) {
		apihandlers.PanicWithMsg("Content dont represent the hash")
	}

	if !ed25519.Verify(signer, hash, sign) {
		apihandlers.PanicWithMsg("Sign dont coincide")
	}

	return t
}
