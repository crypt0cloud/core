package core

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/crypt0cloud/core/crypto"
	md "github.com/crypt0cloud/core/model"
	"github.com/crypt0cloud/core/tools"
	"github.com/onlyangel/apihandlers"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/appengine/log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func block_handlers() {
	BLOCK_CREATION_URL_HANDLER = os.Getenv("BLOCK_CREATION_URL_HANDLER")
	BLOCK_DURATION, _ = time.ParseDuration(os.Getenv("BLOCK_DURATION"))

	http.HandleFunc("/api/setup/initial_blocks", block_insert_initial_blocks)
	http.HandleFunc("/api/v1/block/get_lasts", block_get_last_blocks)
	http.HandleFunc(BLOCK_CREATION_URL_HANDLER, block_calculate_block)

}

var (
	BLOCK_CREATION_URL_HANDLER = os.Getenv("BLOCK_CREATION_URL_HANDLER")
	BLOCK_DURATION             time.Duration
)

func block_calculate_block(w http.ResponseWriter, r *http.Request) {
	//todo: security
	db := model.Open(r, "")
	arr := db.GetLastBlocks(2)

	if arr[0].BlockTime.UnixNano() > time.Now().UnixNano() {
		fmt.Fprintf(w, "Try again latter, Its not the moment for a block calculation")
		return
	}
	ran := rand.New(rand.NewSource(time.Now().UnixNano()))

	//calculation of entire block
	//todo: fixed to to GAE until new migrations

	log.Infof(tools.Context(r), "Blocks: %+v", arr)

	cursor := db.BlockTransactionsCursor(arr[1].Sign)
	h := sha256.New()
	blocksize := 0
	var sha_acum []byte
	sign, finished := db.NextTransactionSign(cursor)
	for !finished {
		sha_acum = h.Sum(sign)
		sign, finished = db.NextTransactionSign(cursor)
		blocksize++
	}
	db.BlockTransactionsCursorClose(cursor)

	//todo: block transactions size

	maxselected := ran.Intn(9) + 1
	position := 0

	cursor = db.BlockTransactionsCursor(arr[0].Sign)
	sign, finished = db.NextTransactionSign(cursor)
	for !finished && position < maxselected {
		sha_acum = h.Sum(sign)
		sign, finished = db.NextTransactionSign(cursor)
		position++
	}
	db.BlockTransactionsCursorClose(cursor)

	block := block_create_block(db, sha_acum, blocksize, position)
	block.BlockTime = block_calculteblocktime(arr[0])
	db.InsertBlock(block)

	tr := block_create_first_block_transaction(db, block)
	db.InsertTransaction(nil, tr)

	jsonstr, err := json.Marshal(block)
	apihandlers.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", string(jsonstr))
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

	block1 := block_create_block(db, tokensha, 0, 0)
	block1.BlockTime = block1.Creation
	db.InsertBlock(block1)

	tr1 := block_create_first_block_transaction(db, block1)
	db.InsertTransaction(nil, tr1)

	token = make([]byte, 128)
	r1.Read(token)
	sha_256 = sha256.New()
	sha_256.Write([]byte(token))
	tokensha = sha_256.Sum(nil)

	block2 := block_create_block(db, tokensha, 0, 0)
	block2.BlockTime = block_calculteblocktime(*block1)
	db.InsertBlock(block2)

	tr2 := block_create_first_block_transaction(db, block2)
	db.InsertTransaction(nil, tr2)

	arr := []md.Block{*block2, *block1}
	jsonstr, err := json.Marshal(arr)
	apihandlers.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", string(jsonstr))
}

func block_calculteblocktime(block md.Block) time.Time {
	return block.BlockTime.Add(BLOCK_DURATION)
}

func block_create_block(db md.ModelDatabase, content []byte, count, nextblockused int) *md.Block {

	myself := db.GetNodeId()
	myself_private := crypto.Base64_decode(myself.PrivateKey)

	signed := ed25519.Sign(myself_private, content)

	block := new(md.Block)
	block.Creation = time.Now()
	block.TransactionsCount = count
	block.NextBlockTransactionsUsed = nextblockused
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
