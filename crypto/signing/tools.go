package signing

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/onlyangel/apihandlers"
	"golang.org/x/crypto/ed25519"

	"github.com/crypt0cloud/core/crypto"
	model "github.com/crypt0cloud/model_go"
)

func SignTransaction(transaction *model.Transaction, masterkey model.MasterKey) {
	sha_256 := sha256.New()

	// calculate payload
	payload, err := json.Marshal(transaction)
	apihandlers.PanicIfNotNil(err)

	// encode (base64 encondign) payload
	transaction.Content = crypto.Base64_encode(payload)

	//calculate hash of payload
	sha_256.Write(payload)
	payload_sha := sha_256.Sum(nil)
	transaction.Hash = crypto.Base64_encode(payload_sha) // store payload hash

	// sign payload with masterkey
	sign := ed25519.Sign(masterkey.CoordinatorPrivate, payload_sha)
	transaction.Sign = crypto.Base64_encode(sign) // store signed payload

	// store masterkey public key
	transaction.Signer = crypto.Base64_encode(masterkey.CoordinatorPublic)
}

func SignBlockRequestTransport(blrq *model.BlockRequestTransport, masterkey *model.MasterKey) {
	sha_256 := sha256.New()

	// calculate payload
	payload, err := json.Marshal(blrq)
	apihandlers.PanicIfNotNil(err)

	// encode (base64 encondign) payload
	blrq.ForInstance.Content = crypto.Base64_encode(payload)

	//calculate hash of payload
	sha_256.Write(payload)
	payload_sha := sha_256.Sum(nil)

	// sign payload with masterkey
	sign := ed25519.Sign(masterkey.CoordinatorPrivate, payload_sha)
	blrq.ForInstance.Sign = crypto.Base64_encode(sign) // store signed payload

	// store masterkey public key
}
