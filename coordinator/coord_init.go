package coordinator

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/crypt0cloud/core/connections"
	"github.com/crypt0cloud/core/crypto"
	"github.com/crypt0cloud/core/crypto/signing"
	"github.com/crypt0cloud/core/tools"
	"github.com/onlyangel/apihandlers"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/appengine/log"
	"math/rand"
	"net/http"
	"time"

	md "github.com/crypt0cloud/core/model"
)

var model md.ModelConnector

func init() {
	var err error
	model, err = md.Open("datastore")
	if err != nil {
		//TODO ERROR
	}
	//http.HandleFunc("/api/api",api_handler)
	http.HandleFunc("/api/coord/register_masterkey", apihandlers.RecoverApi(coord_registerMasterKey))
	http.HandleFunc("/api/coord/register_nodes", apihandlers.RecoverApi(coord_registerNewNode))
	http.HandleFunc("/api/coord/verify_with_peers", apihandlers.RecoverApi(coord_verifyWithPeers))
	http.HandleFunc("/api/v1/coord/add_app", apihandlers.RecoverApi(coord_addApp))

	//TODO: change method and WS
	http.HandleFunc("/api/v1/coord/scale_with_peers", apihandlers.RecoverApi(coord_scaleWithPeers))
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

	bodydecoder := json.NewDecoder(r.Body)

	cn := new(struct {
		Content string
		Sign    string
	})
	err := bodydecoder.Decode(cn)
	defer r.Body.Close()
	apihandlers.PanicIfNotNil(err)

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

	for _, url := range nodesdata.Urls {
		nodeID := connections.GetRemoteNodeCredentials(r, url)

		sha_256 := sha256.New()

		transaction := new(md.Transaction)
		transaction.SignerKinds = []string{"__REGISTERNODE"}
		transaction.SignKind = "__REGISTERNODE"
		transaction.AppID = base64.StdEncoding.EncodeToString(mk.CoordinatorPublic)
		transaction.Parent = ""
		transaction.Callback = "http://" + mk.URL
		transaction.Payload = url

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

		traurl := "http://" + url + "/api/v1/post_single_transaction"
		response := connections.PostRemote(r, traurl, jsonstr)

		log.Debugf(ctx, "Transaction response from: '%s'", traurl)
		log.Debugf(ctx, string(response))

		err = json.Unmarshal(response, transaction)
		apihandlers.PanicIfNotNil(err)

		db.Coord_Insert_ExternalNode(nodeID)
	}
}

func coord_verifyWithPeers(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")
	arr := db.Coord_GetRandomNodeIdentification(1)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	bts := buf.Bytes()

	sino := true

	for _, node := range arr {
		response := connections.PostRemote(r, "http://"+node.Endpoint+"/api/v1/pair_verification", bts)
		responsestr := string(response)

		error := new(apihandlers.ErrorType)
		err := json.Unmarshal(response, error)
		apihandlers.PanicIfNotNil(err)

		if error.Error != "" {
			apihandlers.PanicWithMsg(responsestr)
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

func coord_addApp(w http.ResponseWriter, r *http.Request) {
	db := model.Open(r, "")

	//Validate criptographically transaction body
	t := crypto.Validate_criptoTransaction(r.Body)

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

		transaction := new(md.Transaction)
		transaction.SignerKinds = []string{"__NEWAPP"}
		transaction.SignKind = "__NEWAPP"
		transaction.AppID = t.AppID
		transaction.Parent = ""
		transaction.Callback = t.Callback
		transaction.Payload = "__NEWAPP"

		transaction.ToNode = node
		transaction.FromNode = *mySelf

		signing.SignTransaction(transaction, *masterkey)

		jsonstr, err := json.Marshal(transaction)
		apihandlers.PanicIfNotNil(err)

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
