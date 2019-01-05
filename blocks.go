package core

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/crypt0cloud/core/crypto"
	md "github.com/crypt0cloud/core/model"
	"github.com/onlyangel/apihandlers"
	"golang.org/x/crypto/ed25519"
	"math/rand"
	"net/http"
	"time"
)

func block_handlers() {
	http.HandleFunc("/api/setup/initial_blocks", block_insert_initial_blocks)
	http.HandleFunc("/api/v1/block/get_lasts", block_get_last_blocks)
}

func block_get_last_blocks(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	arr := db.GetLastBlocks(2)
	jsonstr, err := json.Marshal(arr)
	apihandlers.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", string(jsonstr))
}

func block_insert_initial_blocks(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	if count := db.CountBlocks(); count > 0 {
		apihandlers.PanicWithMsg("Already initted blocks")
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	token := make([]byte, 128)
	r1.Read(token)
	sha_256 := sha256.New()
	sha_256.Write([]byte(token))
	tokensha := sha_256.Sum(nil)

	block1 := block_create_block(db, tokensha, 0)
	db.InsertBlock(block1)

	tr1 := block_create_first_block_transaction(db, block1)
	db.InsertTransaction(nil, tr1)

	token = make([]byte, 128)
	r1.Read(token)
	sha_256 = sha256.New()
	sha_256.Write([]byte(token))
	tokensha = sha_256.Sum(nil)

	block2 := block_create_block(db, tokensha, 0)
	db.InsertBlock(block2)

	tr2 := block_create_first_block_transaction(db, block2)
	db.InsertTransaction(nil, tr2)

	arr := []md.Block{*block2, *block1}
	jsonstr, err := json.Marshal(arr)
	apihandlers.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", string(jsonstr))
}

func block_create_block(db md.ModelDatabase, content []byte, count int) *md.Block {

	myself := db.GetNodeId()
	myself_private := crypto.Base64_decode(myself.PrivateKey)

	signed := ed25519.Sign(myself_private, content)

	block := new(md.Block)
	block.Creation = time.Now()
	block.TransactionsCount = count
	block.Hash = crypto.Base64_encode(content)
	block.Sign = crypto.Base64_encode(signed)

	return block
}

func block_create_first_block_transaction(db md.ModelDatabase, block *md.Block) *md.Transaction {
	myself := db.GetNodeId()
	myself_private := crypto.Base64_decode(myself.PrivateKey)
	myself.PrivateKey = ""

	tr := new(md.Transaction)
	tr.Creation = time.Now().UnixNano()
	tr.Payload = "Block Transaction"
	tr.SignKind = "_BLOCKTRANSACTION"
	tr.SignerKinds = []string{"_BLOCKTRANSACTION"}
	tr.FromNode = *myself
	tr.BlockSign = block.Sign

	jsonstr, err := json.Marshal(tr)
	apihandlers.PanicIfNotNil(err)

	tr.Content = base64.StdEncoding.EncodeToString(jsonstr)

	sha_256 := sha256.New()
	sha_256.Write(jsonstr)
	contentsha := sha_256.Sum(nil)
	tr.Hash = base64.StdEncoding.EncodeToString(contentsha)

	sign := ed25519.Sign(myself_private, contentsha)
	tr.Sign = base64.StdEncoding.EncodeToString(sign)

	tr.Signer = myself.PublicKey

	return tr
}
