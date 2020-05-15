package coordinator

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/onlyangel/apihandlers"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/crypt0cloud/core/connections"
	"github.com/crypt0cloud/core/tools"
	crypto "github.com/crypt0cloud/crypto_go"
	"github.com/crypt0cloud/crypto_go/signing"

	md "github.com/crypt0cloud/model_go"
)

var model md.ModelConnector

var BLOCK_DURATION time.Duration

func init() {
	var err error
	model, err = md.Open("datastore")
	if err != nil {
		//TODO ERROR
	}
	//http.HandleFunc("/api/api",api_handler)
	http.HandleFunc("/api/v1/coord/register_masterkey", apihandlers.RecoverApi(coord_registerMasterKey))
	http.HandleFunc("/api/v1/coord/register_nodes", apihandlers.RecoverApi(coord_registerNewNode))
	http.HandleFunc("/api/v1/coord/verify_with_peers", apihandlers.RecoverApi(coord_verifyWithPeers))
	http.HandleFunc("/api/v1/coord/add_app", apihandlers.RecoverApi(coord_addApp))

	//TODO: change method and WS
	http.HandleFunc("/api/v1/coord/scale_with_peers", apihandlers.RecoverApi(coord_scaleWithPeers))

	http.HandleFunc("/api/v1/coord/blockcalculationauthorization", apihandlers.RecoverApi(coord_handleBlockCalculationAuthorization))

	http.HandleFunc("/api/v1/coord/nodeBlockSigning", apihandlers.RecoverApi(coord_handleNodeBlockSigning))

	BLOCK_DURATION, _ = time.ParseDuration("5m")
}

func coord_handleNodeBlockSigning(w http.ResponseWriter, r *http.Request) {
	bodydecoder := json.NewDecoder(r.Body)

	t := new(md.BlockRequestForInstance)
	err := bodydecoder.Decode(t)
	apihandlers.PanicIfNotNil(err)
	defer r.Body.Close()

	ctx := appengine.NewContext(r)
	log.Infof(ctx, "BLOCK TO SIGN: %+v", t)

	db := model.Open(r, "")
	localt := db.Coord_GetLocalBlockRequestForInstance(t)

	sha_acum := crypto.Base64_decode(t.BlockSignProposal)
	signed := crypto.Base64_decode(t.BlockSignProposalSign)
	pkey := crypto.Base64_decode(localt.NodePubK)
	autentic := ed25519.Verify(pkey, sha_acum, signed)

	if !autentic {
		apihandlers.PanicWithMsg("Non authoriced")
	}

	coordKey := db.Coord_GetKey()

	signedblock := ed25519.Sign(coordKey.CoordinatorPrivate, sha_acum)
	t.BlockSign = crypto.Base64_encode(signedblock)

	db.Coord_UpdateBlockRequestForInstance(t)

	req := &md.BlockRequestTransport{}
	content := crypto.Base64_decode(t.Content)
	err = json.Unmarshal(content, req)
	apihandlers.PanicIfNotNil(err)

	block := &md.Block{}
	block.Hash = t.BlockSignProposal
	block.TransactionsCount = t.BlockSignProposalCount
	block.NextBlockTransactionsUsed = req.Request.SeedTrCount
	block.Sign = t.BlockSign

	block.Creation = time.Unix(0, req.Request.TimeCreation)
	block.BlockTime = time.Unix(0, req.Request.TimeValidity)

	jsonstr, err := json.Marshal(block)
	apihandlers.PanicIfNotNil(err)
	connections.PostRemote(r, fmt.Sprintf("http://%s/api/internal/signed_block_insert", t.Endpoint), jsonstr)
}

func coord_handleBlockCalculationAuthorization(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	key_exists := db.Coord_MasterKeyExists()

	if !key_exists {
		apihandlers.PanicWithMsg("Coordinator need to be initted")
	}

	coordKey := db.Coord_GetKey()
	moment := time.Now()
	moment_creation_int := moment.UnixNano()

	halfblockduration, _ := time.ParseDuration("2.5m")
	moment_of_validity := moment.Add(halfblockduration)
	moment_of_validity_int := moment_of_validity.UnixNano()

	ran := rand.New(rand.NewSource(time.Now().UnixNano()))

	blreq := &md.BlockRequest{
		0,
		moment_creation_int,
		moment_of_validity_int,
		ran.Intn(9),
	}

	db.Coord_InsertBlockRequest(blreq)

	arr := db.Coord_GetRandomNodeIdentification(0)

	ctx := tools.Context(r)
	for _, node := range arr {
		fi := &md.BlockRequestForInstance{
			BlockRequest: blreq.IdVal,
			Nonce:        ran.Int63(),
			Endpoint:     node.Endpoint,
			NodePubK:     node.PublicKey,
		}

		brt := &md.BlockRequestTransport{
			Request:     blreq,
			ForInstance: fi,
		}

		signing.SignBlockRequestTransport(brt, coordKey)
		db.Coord_InsertBlockRequestForInstance(fi)

		jsonstr, _ := json.Marshal(brt)
		url := fmt.Sprintf("http://%s/api/internal/calculate_block", node.Endpoint)
		connections.PostRemote(r, url, jsonstr)
		fmt.Fprint(w, string(jsonstr))

		log.Infof(ctx, "\n\n\n\nENVIANDO A: %s\n%s\n", url, string(jsonstr))
	}

	// generate authentication

	// get all nodes

	// send authentication to nodes
}

func coord_registerMasterKey(w http.ResponseWriter, r *http.Request) {
	//ctx := tools.Context(r)
	db := model.Open(r, "")
	key_exists := db.Coord_MasterKeyExists()

	if key_exists {
		apihandlers.PanicWithMsg("Already asigned key")
	}

	key := tools.FormValueEscaped(r, "key")
	url := tools.FormValueEscaped(r, "url")
	if key == "" || url == "" {
		apihandlers.PanicWithMsg("In parameters")
	}

	keyarr, err := base64.StdEncoding.DecodeString(key)
	apihandlers.PanicIfNotNil(err)

	makey := new(md.MasterKey)
	makey.MasterPublicKey = keyarr
	makey.URL = url
	makey.CoordinatorPublic, makey.CoordinatorPrivate, err = ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))

	db.Coord_InsertKey(makey)

	tools.PrintJson(w, makey)
}

func coord_registerNewNode(w http.ResponseWriter, r *http.Request) {
	ctx := tools.Context(r)
	db := model.Open(r, "")
	mk := db.Coord_GetKey()
	myNodeID := db.GetNodeId()

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	cn := new(struct {
		Content string
		Sign    string
	})

	err = json.Unmarshal(body, cn)
	apihandlers.PanicIfNotNil(err)

	log.Infof(ctx, "%+v", cn)

	sign, err := base64.StdEncoding.DecodeString(cn.Sign)
	apihandlers.PanicIfNotNil(err)

	content, err := base64.StdEncoding.DecodeString(cn.Content)
	apihandlers.PanicIfNotNil(err)

	sha_256 := sha256.New()
	sha_256.Write(content)
	contentsha := sha_256.Sum(nil)

	if !ed25519.Verify(mk.MasterPublicKey, contentsha, sign) {
		apihandlers.PanicWithMsg("verification of sign problem")
	}

	nodesdata := new(struct {
		Urls []string
	})

	err = json.Unmarshal(content, nodesdata)
	apihandlers.PanicIfNotNil(err)

	genesis := get_GenesisBlocks(r)
	genesisjson, _ := json.Marshal(genesis)

	for _, url := range nodesdata.Urls {

		nodeID := connections.GetRemoteNodeCredentials(r, url)

		blockurl := "http://" + url + "/api/setup/initial_blocks"
		response := connections.PostRemote(r, blockurl, genesisjson)
		var arr []md.Block
		err = json.Unmarshal(response, &arr)
		apihandlers.PanicIfNotNil(err)

		log.Infof(ctx, "NODEARR: %+v", arr[0])

		err = json.Unmarshal(response, &arr)
		apihandlers.PanicIfNotNil(err)

		sha_256 := sha256.New()

		transaction := new(md.Transaction)
		transaction.SignerKinds = []string{"__REGISTERNODE"}
		transaction.SignKind = "__REGISTERNODE"
		transaction.AppID = base64.StdEncoding.EncodeToString(mk.CoordinatorPublic)
		transaction.Parent = ""
		transaction.Callback = "http://" + mk.URL
		transaction.Payload = url
		transaction.Creation = time.Now().UnixNano()
		transaction.BlockSign = arr[0].Sign

		transaction.ToNode = *nodeID
		transaction.FromNode = *myNodeID

		jsonstr, err := json.Marshal(transaction)
		apihandlers.PanicIfNotNil(err)

		transaction.Content = base64.StdEncoding.EncodeToString(jsonstr)
		sha_256.Write(jsonstr)
		contentsha := sha_256.Sum(nil)
		transaction.Hash = base64.StdEncoding.EncodeToString(contentsha)

		sign := ed25519.Sign(mk.CoordinatorPrivate, contentsha)
		transaction.Sign = base64.StdEncoding.EncodeToString(sign)

		transaction.Signer = transaction.AppID

		jsonstr, err = json.Marshal(transaction)
		apihandlers.PanicIfNotNil(err)

		traurl := "http://" + url + "/api/setup/set_initial_node_registration"
		response = connections.PostRemote(r, traurl, jsonstr)

		log.Debugf(ctx, "Transaction response from: '%s'", traurl)
		log.Debugf(ctx, string(response))

		err = json.Unmarshal(response, transaction)
		apihandlers.PanicIfNotNil(err)

		db.Coord_Insert_ExternalNode(nodeID)
	}
}

func get_GenesisBlocks(r *http.Request) *md.GenesisBlocksTransport {
	db := model.Open(r, "")

	if count := db.CountBlocks(); count > 0 {
		return db.Coord_GetGenesisBlocks()
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	token := make([]byte, 128)
	r1.Read(token)
	sha_256 := sha256.New()
	sha_256.Write(token)
	tokensha := sha_256.Sum(nil)

	block1 := block_create_block(db, tokensha, 0, 0)
	block1.BlockTime = block1.Creation
	db.InsertBlock(block1)

	tr1 := block_create_first_block_transaction(db, block1)
	db.InsertTransaction(nil, tr1)

	token = make([]byte, 128)
	r1.Read(token)
	sha_256 = sha256.New()
	sha_256.Write(token)
	tokensha = sha_256.Sum(nil)

	block2 := block_create_block(db, tokensha, 0, 0)
	block2.BlockTime = block_calculteblocktime(*block1)
	db.InsertBlock(block2)

	tr2 := block_create_first_block_transaction(db, block2)
	db.InsertTransaction(nil, tr2)

	fbt := new(md.GenesisBlocksTransport)
	fbt.Block1 = block1
	fbt.Block2 = block2
	fbt.Transaction11 = tr1
	fbt.Transaction21 = tr2

	db.Coord_SetGenesisBlocks(fbt)

	return fbt
}
func block_create_block(db md.ModelDatabase, content []byte, count, nextblockused int) *md.Block {

	myself := db.Coord_GetKey()

	signed := ed25519.Sign(myself.CoordinatorPrivate, content)

	block := new(md.Block)
	block.Creation = time.Now()
	block.TransactionsCount = count
	block.NextBlockTransactionsUsed = nextblockused
	block.Hash = crypto.Base64_encode(content)
	block.Sign = crypto.Base64_encode(signed)

	return block
}
func block_calculteblocktime(block md.Block) time.Time {
	return block.BlockTime.Add(BLOCK_DURATION)
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

func coord_verifyWithPeers(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	arr := db.Coord_GetRandomNodeIdentification(0)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	bts := buf.Bytes()

	sino := true

	for _, node := range arr {
		response := connections.PostRemote(r, "http://"+node.Endpoint+"/api/v1/pair_verification", bts)
		responsestr := string(response)

		if responsestr == "false" {
			sino = false
		} else if responsestr == "true" {
			sino = true
		} else {
			error := new(apihandlers.ErrorType)
			err := json.Unmarshal(response, error)
			apihandlers.PanicIfNotNil(err)

			if error.Error != "" {
				apihandlers.PanicWithMsg(responsestr)
			}
		}

	}

	if sino {
		fmt.Fprintf(w, "true")
	} else {
		fmt.Fprintf(w, "false")
	}

}

func coord_addApp(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	//Validate criptographically transaction body
	t := crypto.Validate_criptoTransaction(r.Body)

	log.Infof(tools.Context(r), "REceiving transaction order: %+v", t)

	// Validate transaction data
	if t.SignKind != "NewApp" {
		apihandlers.PanicWithMsg("No New App transaction")
	}

	if t.Payload == "" {
		apihandlers.PanicWithMsg("Error in sign credentials order")
	}
	if len(t.SignerKinds) != 1 {
		apihandlers.PanicWithMsg("Single transaction should have a parent = 0")
	}

	if t.SignKind != t.SignerKinds[0] {
		apihandlers.PanicWithMsg("Single transaction should have a parent = 0")
	}

	//todo verify the callback it is a url
	if t.Callback == "" {
		apihandlers.PanicWithMsg("There should be a callback")
	}

	if db.AppIdExists(r, t.AppID) {
		apihandlers.PanicWithMsg("App sign already exists")
	}

	// Get get Keys
	masterkey := db.Coord_GetKey()

	// Get all nodes
	arr := db.Coord_GetRandomNodeIdentification(0)

	// get My node
	mySelf := db.GetNodeId()

	// Create the app in all the nodes
	for _, node := range arr {

		log.Infof(tools.Context(r), "%+v", node)

		resp, err := connections.CallRemote(r, fmt.Sprintf("http://%s/api/v1/block/get_lasts", node.Endpoint))
		apihandlers.PanicIfNotNil(err)
		var arr []md.Block
		err = json.Unmarshal(resp, &arr)
		apihandlers.PanicIfNotNil(err)

		transaction := new(md.Transaction)
		transaction.SignerKinds = []string{"__NEWAPP"}
		transaction.SignKind = "__NEWAPP"
		transaction.AppID = t.AppID
		transaction.Parent = ""
		transaction.Callback = t.Callback
		transaction.Payload = "__NEWAPP"
		transaction.BlockSign = arr[0].Sign

		transaction.Creation = time.Now().UnixNano()

		transaction.ToNode = node
		transaction.FromNode = *mySelf

		signing.SignTransaction(transaction, *masterkey)

		jsonstr, err := json.Marshal(transaction)
		apihandlers.PanicIfNotNil(err)

		log.Infof(tools.Context(r), "Adding app: %+v", transaction)

		response := connections.PostRemote(r, "http://"+node.Endpoint+"/api/v1/post_single_transaction", jsonstr)

		err = json.Unmarshal(response, transaction)
		apihandlers.PanicIfNotNil(err)

	}

	fmt.Fprintf(w, "OK")

}

func coord_scaleWithPeers(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	crypto.Validate_criptoTransaction(r.Body)

	arr := db.Coord_GetRandomNodeIdentification(1)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	bts := buf.Bytes()

	sino := true

	for _, node := range arr {
		response := connections.PostRemote(r, "http://"+node.Endpoint+"/api/v1/pair_verification", bts)
		responsestr := string(response)

		if haserror, errorstr := tools.API_Error(response); haserror {
			apihandlers.PanicWithMsg(errorstr)
		}

		if responsestr == "false" {
			sino = false
		}

	}

	if sino {
		fmt.Fprintf(w, "true")
	} else {
		fmt.Fprintf(w, "false")
	}

}
