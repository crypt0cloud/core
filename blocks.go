package core

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	api "github.com/onlyangel/apihandlers"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/appengine/log"

	"github.com/crypt0cloud/core/connections"
	"github.com/crypt0cloud/core/tools"
	crypto "github.com/crypt0cloud/crypto_go"
	md "github.com/crypt0cloud/model_go"
)

func block_handlers() {
	BLOCK_CREATION_URL_HANDLER = os.Getenv("BLOCK_CREATION_URL_HANDLER")
	BLOCK_DURATION, _ = time.ParseDuration(os.Getenv("BLOCK_DURATION"))

	http.HandleFunc("/api/setup/initial_blocks", block_insert_initial_blocks)
	http.HandleFunc("/api/v1/block/get_lasts", block_get_last_blocks)
	//http.HandleFunc(BLOCK_CREATION_URL_HANDLER, block_calculate_block)
	http.HandleFunc("/api/internal/calculate_block", block_calculate_block)
	http.HandleFunc("/api/internal/signed_block_insert", block_signed_block_insert)
}

var (
	BLOCK_CREATION_URL_HANDLER = os.Getenv("BLOCK_CREATION_URL_HANDLER")
	BLOCK_DURATION             time.Duration
)

func block_signed_block_insert(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	block := &md.Block{}

	err = json.Unmarshal(body, block)
	api.PanicIfNotNil(err)

	db := model.Open(r, "")

	brfi := db.GetBlockRequestForInstanceFromHash(block.Hash)
	tr := db.GetCoordinatorKey(r)
	masterkye := crypto.Base64_decode(tr.AppID)
	sign := crypto.Base64_decode(block.Sign)
	hash := crypto.Base64_decode(brfi.BlockSignProposal)

	valid := ed25519.Verify(masterkye, hash, sign)

	if !valid {
		api.PanicWithMsg("NOT VALID")
	}

	db.InsertBlock(block)

	tra := block_create_first_block_transaction(db, block)
	db.InsertTransaction(r, tra)

	fmt.Fprintf(w, "OK")
}

func block_calculate_block(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	arr := db.GetLastBlocks(2)

	if arr[0].BlockTime.UnixNano() > time.Now().UnixNano() {
		api.PanicWithMsg("Try again latter, Its not the moment for a block calculation")
		return
	}

	coordinatorkey := db.GetCoordinatorKey(r)

	var brt *md.BlockRequestTransport
	if brt_, valid := crypto.Validate_blockRequestTransport(r.Body, coordinatorkey); !valid {
		api.PanicWithMsg("Not authoriced")
	} else {
		brt = brt_
	}

	db.InsertBlockRequest(brt.Request)
	db.InsertBlockRequestForInstance(brt.ForInstance)

	log.Infof(tools.Context(r), "Blocks: %+v", arr)

	// BEGIN calculation of block hash
	//TODO: migrate block hach calculation to a background service
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

	position := 0

	cursor = db.BlockTransactionsCursor(arr[0].Sign)
	sign, finished = db.NextTransactionSign(cursor)
	for !finished && position < brt.Request.SeedTrCount {
		sha_acum = h.Sum(sign)
		sign, finished = db.NextTransactionSign(cursor)
		position++
	}
	db.BlockTransactionsCursorClose(cursor)
	// END calculation of block hash
	myself := db.GetNodeId()
	myself_private := crypto.Base64_decode(myself.PrivateKey)

	signed := ed25519.Sign(myself_private, sha_acum)
	brt.ForInstance.BlockSignProposal = crypto.Base64_encode(sha_acum)
	brt.ForInstance.BlockSignProposalSign = crypto.Base64_encode(signed)
	brt.ForInstance.BlockSignProposalCount = blocksize

	jsonstr, err := json.Marshal(brt.ForInstance)
	api.PanicIfNotNil(err)

	db.UpdateBlockRequestForInstance(brt.ForInstance)

	url := fmt.Sprintf("%s/api/v1/coord/nodeBlockSigning", coordinatorkey.Callback)
	log.Infof(tools.Context(r), "\n\n\n\nENVIANDO A: %s\n%s\n", url, string(jsonstr))
	connections.PostRemote(r, url, jsonstr)

	return
}

func block_get_last_blocks(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	arr := db.GetLastBlocks(2)
	jsonstr, err := json.Marshal(arr)
	if err != nil {
		// lsdkflksdjfdj
	}
	api.PanicIfNotNil(err)

	fmt.Fprintf(w, "%s", string(jsonstr))
}

func block_insert_initial_blocks(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	if count := db.CountBlocks(); count > 0 {
		api.PanicWithMsg("Already initted blocks")
	}

	tra := new(md.GenesisBlocksTransport)

	err := json.NewDecoder(r.Body).Decode(tra)
	api.PanicIfNotNil(err)

	db.InsertBlock(tra.Block1)
	db.InsertTransaction(r, tra.Transaction11)

	db.InsertBlock(tra.Block2)
	db.InsertTransaction(r, tra.Transaction21)

	arr := []md.Block{*tra.Block2, *tra.Block1}
	jsonstr, err := json.Marshal(arr)
	api.PanicIfNotNil(err)

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
	api.PanicIfNotNil(err)

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
