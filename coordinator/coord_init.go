package coordinator

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/crypt0cloud/core/connections"
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
	apihandlers.PanicIfNil(err)

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
	apihandlers.PanicIfNil(err)

	sign, err := base64.StdEncoding.DecodeString(cn.Sign)
	apihandlers.PanicIfNil(err)

	content, err := base64.StdEncoding.DecodeString(cn.Content)
	apihandlers.PanicIfNil(err)

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
	apihandlers.PanicIfNil(err)

	for _, url := range nodesdata.Urls {
		nodeID := connections.GetRemoteNodeCredentials(r, url)

		sha_256 := sha256.New()

		transaction := new(md.Transaction)
		transaction.SignerKinds = []string{"__REGISTERNODE"}
		transaction.SignKind = "__REGISTERNODE"
		transaction.AppID = base64.StdEncoding.EncodeToString(mk.CoordinatorPublic)
		transaction.Parent = 0
		transaction.Callback = "http://" + mk.URL
		transaction.Payload = url

		transaction.ToNode = *nodeID
		transaction.FromNode = *myNodeID

		jsonstr, err := json.Marshal(transaction)
		apihandlers.PanicIfNil(err)

		transaction.Content = base64.StdEncoding.EncodeToString(jsonstr)
		sha_256.Write(jsonstr)
		contentsha := sha_256.Sum(nil)
		transaction.Hash = base64.StdEncoding.EncodeToString(contentsha)

		sign := ed25519.Sign(mk.CoordinatorPrivate, contentsha)
		transaction.Sign = base64.StdEncoding.EncodeToString(sign)

		transaction.Signer = transaction.AppID

		jsonstr, err = json.Marshal(transaction)
		apihandlers.PanicIfNil(err)

		traurl := "http://" + url + "/api/v1/post_single_transaction"
		response := connections.PostRemote(r, traurl, jsonstr)

		log.Debugf(ctx, "Transaction response from: '%s'", traurl)
		log.Debugf(ctx, string(response))

		err = json.Unmarshal(response, transaction)
		apihandlers.PanicIfNil(err)

		db.Coord_Insert_ExternalNode(nodeID)
	}
}
