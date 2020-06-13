package model_gae

import (
	"net/http"
	"strconv"
	"time"

	api "github.com/onlyangel/apihandlers"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/crypt0cloud/core/tools"
	crypto "github.com/crypt0cloud/crypto_go"
	model "github.com/crypt0cloud/model_go"
)

func init() {
	model.RegisterModel("datastore", &ConnectorDatastore{})
}

type ConnectorDatastore struct {
	model.ModelDatabase
}

func (d ConnectorDatastore) Open(r *http.Request, config string) model.ModelDatabase {
	return DatabaseDatastore{
		ctx: tools.Context(r),
	}
}

type DatabaseDatastore struct {
	ctx context.Context
}

/**
Coordinator
*/
func (d DatabaseDatastore) Coord_MasterKeyExists() bool {
	q := datastore.NewQuery("MasterKey")
	count, err := q.Count(d.ctx)
	api.PanicIfNotNil(err)

	return count > 0
}

func (d DatabaseDatastore) Coord_InsertKey(key *model.MasterKey) {
	k := datastore.NewIncompleteKey(d.ctx, "MasterKey", nil)
	k, err := datastore.Put(d.ctx, k, key)
	api.PanicIfNotNil(err)
}

func (d DatabaseDatastore) Coord_GetKey() *model.MasterKey {
	q := datastore.NewQuery("MasterKey")
	mk := new(model.MasterKey)

	if count, err := q.Count(d.ctx); err != nil {
		api.PanicIfNotNil(err)
	} else if count == 0 {
		api.PanicWithMsg("Not initialiced Coordinator")
	}

	t := q.Run(d.ctx)
	_, err := t.Next(mk)
	api.PanicIfNotNil(err)
	return mk
}

func (d DatabaseDatastore) Coord_Insert_ExternalNode(node *model.NodeIdentification) {
	k := datastore.NewIncompleteKey(d.ctx, "ExternalNodesIdentification", nil)
	_, err := datastore.Put(d.ctx, k, node)
	api.PanicIfNotNil(err)
}

func (d DatabaseDatastore) Coord_GetRandomNodeIdentification(limit int) []model.NodeIdentification {
	q := datastore.NewQuery("ExternalNodesIdentification")
	if limit > 0 {
		q = q.Limit(limit) //Limitar a uno
	}

	var arr []model.NodeIdentification
	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)

	if len(arr) == 0 {
		api.PanicWithMsg("No Nodes")
	}

	return arr
}

func (d DatabaseDatastore) Coord_InsertBlockRequest(blreq *model.BlockRequest) {
	k := datastore.NewIncompleteKey(d.ctx, "Coord.BlockRequest", nil)
	k, err := datastore.Put(d.ctx, k, blreq)
	api.PanicIfNotNil(err)

	blreq.IdVal = k.IntID()
	_, err = datastore.Put(d.ctx, k, blreq)
	api.PanicIfNotNil(err)
}

func (d DatabaseDatastore) Coord_InsertBlockRequestForInstance(transport *model.BlockRequestForInstance) {
	k := datastore.NewIncompleteKey(d.ctx, "Coord.BlockRequestForInstance", nil)
	k, err := datastore.Put(d.ctx, k, transport)
	api.PanicIfNotNil(err)
}

func (d DatabaseDatastore) Coord_GetLocalBlockRequestForInstance(transport *model.BlockRequestForInstance) *model.BlockRequestForInstance {
	q := datastore.NewQuery("Coord.BlockRequestForInstance").Filter("NodePubK = ", transport.NodePubK).Filter("Nonce = ", transport.Nonce)
	count, err := q.Count(d.ctx)
	api.PanicIfNotNil(err)

	if count == 0 {
		return nil
	}

	tra := &model.BlockRequestForInstance{}
	ti := q.Run(d.ctx)
	_, err = ti.Next(tra)
	api.PanicIfNotNil(err)

	return tra
}

func (d DatabaseDatastore) Coord_UpdateBlockRequestForInstance(transport *model.BlockRequestForInstance) {
	q := datastore.NewQuery("Coord.BlockRequestForInstance").Filter("NodePubK = ", transport.NodePubK).Filter("Nonce = ", transport.Nonce).Limit(1).KeysOnly()
	count, err := q.Count(d.ctx)
	api.PanicIfNotNil(err)

	if count == 0 {
		api.PanicWithMsg("Invalid Transaction")
	}

	tra := &model.BlockRequestForInstance{}
	ti := q.Run(d.ctx)
	k, err := ti.Next(tra)
	api.PanicIfNotNil(err)

	k, err = datastore.Put(d.ctx, k, transport)
	api.PanicIfNotNil(err)
}

func (d DatabaseDatastore) Coord_SetGenesisBlocks(genesis *model.GenesisBlocksTransport) {
	k := datastore.NewIncompleteKey(d.ctx, "Coord.Block", nil)
	_, err := datastore.Put(d.ctx, k, genesis.Block1)
	api.PanicIfNotNil(err)

	k = datastore.NewIncompleteKey(d.ctx, "Coord.Block", nil)
	_, err = datastore.Put(d.ctx, k, genesis.Block2)
	api.PanicIfNotNil(err)

	k = datastore.NewIncompleteKey(d.ctx, "Coord.Transaction", nil)
	_, err = datastore.Put(d.ctx, k, genesis.Transaction11)
	api.PanicIfNotNil(err)

	k = datastore.NewIncompleteKey(d.ctx, "Coord.Transaction", nil)
	_, err = datastore.Put(d.ctx, k, genesis.Transaction21)
	api.PanicIfNotNil(err)
}

func (d DatabaseDatastore) Coord_GetGenesisBlocks() *model.GenesisBlocksTransport {
	q := datastore.NewQuery("Coord.Block").Order("BlockTime").Limit(2)
	var blocks []model.Block
	_, err := q.GetAll(d.ctx, &blocks)
	api.PanicIfNotNil(err)

	q = datastore.NewQuery("Coord.Block").Order("BlockTime").Limit(2)
	var trs []model.Transaction
	_, err = q.GetAll(d.ctx, &trs)
	api.PanicIfNotNil(err)

	gen := new(model.GenesisBlocksTransport)
	gen.Block1 = &blocks[0]
	gen.Block2 = &blocks[1]

	gen.Transaction11 = &trs[0]
	gen.Transaction21 = &trs[1]

	return gen
}

/**
Nodes
*/
func (d DatabaseDatastore) IsRegisteredNodeID() bool {
	q := datastore.NewQuery("NodeIdentification")
	count, err := q.Count(d.ctx)
	api.PanicIfNotNil(err)

	return count > 0
}

func (d DatabaseDatastore) RegisteredNodeID(us *model.NodeIdentification) {
	k := datastore.NewIncompleteKey(d.ctx, "NodeIdentification", nil)
	k, err := datastore.Put(d.ctx, k, us)
	api.PanicIfNotNil(err)
}

func (d DatabaseDatastore) GetNodeId() *model.NodeIdentification {
	q := datastore.NewQuery("NodeIdentification").Filter("Myself =", true)
	var arr []model.NodeIdentification
	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)
	if len(arr) == 0 {
		api.PanicWithMsg("Not Innited Node")
	}
	return &arr[0]
}

func (d DatabaseDatastore) SetupSetEndPointIfNull(endpoint string) bool {
	q := datastore.NewQuery("NodeIdentification")
	var arr []model.NodeIdentification
	k, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)
	if len(arr) == 0 {
		api.PanicWithMsg("Not Innited Node")
	}

	if arr[0].Endpoint != "" {
		return false
	}

	arr[0].Endpoint = endpoint
	_, err = datastore.Put(d.ctx, k[0], &arr[0])
	api.PanicIfNotNil(err)

	return true
}

func (d DatabaseDatastore) SetupNodeRegistrationInDeployment(t *model.Transaction) {
	k := datastore.NewIncompleteKey(d.ctx, "RegisterNode.Transaction", nil)
	t.InsertMoment = time.Now().UnixNano()
	k, err := datastore.Put(d.ctx, k, t)
	api.PanicIfNotNil(err)

	t.IdVal = k.IntID()
	k, err = datastore.Put(d.ctx, k, t)
	api.PanicIfNotNil(err)

}

/*
Apps
*/

func (d DatabaseDatastore) AppIdExists(r *http.Request, id string) bool {
	q := datastore.NewQuery("Transaction").Filter("AppID = ", id).Limit(1)
	count, err := q.Count(d.ctx)
	api.PanicIfNotNil(err)

	return count > 0
}

/*
Ledger
*/

func (d DatabaseDatastore) NodeExists(r *http.Request) bool {

	q := datastore.NewQuery("RegisterNode.Transaction").Filter("SignKind = ", "__REGISTERNODE").Limit(1)
	count, err := q.Count(d.ctx)
	api.PanicIfNotNil(err)

	return count > 0
}

func (d DatabaseDatastore) GetCoordinatorKey(r *http.Request) *model.Transaction {

	q := datastore.NewQuery("RegisterNode.Transaction").Filter("SignKind = ", "__REGISTERNODE")
	var arr []model.Transaction
	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)

	if len(arr) == 0 {
		api.PanicWithMsg("Not Registered Node")
	}
	return &arr[0]
}

func (d DatabaseDatastore) UserSignExist(r *http.Request, id string) *model.Transaction {
	q := datastore.NewQuery("Transaction").Filter("SignKind = ", "NewUser").Filter("Signer = ", id).Limit(1)

	var arr []model.Transaction
	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)

	if len(arr) == 0 {
		return nil
	} else {
		return &arr[0]
	}
}

func (d DatabaseDatastore) UserPayloadExist(r *http.Request, id string) *model.Transaction {
	q := datastore.NewQuery("Transaction").Filter("SignKind = ", "NewUser").Filter("Payload = ", id).Limit(1)

	var arr []model.Transaction
	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)

	if len(arr) == 0 {
		return nil
	} else {
		return &arr[0]
	}

}

/*
	Transactions
*/

func (d DatabaseDatastore) InsertPairVerificationTransaction(r *http.Request, t *model.Transaction) *model.Transaction {
	k := datastore.NewIncompleteKey(d.ctx, "Transaction", nil)
	k, err := datastore.Put(d.ctx, k, t)
	api.PanicIfNotNil(err)

	return t
}

func (d DatabaseDatastore) InsertTransaction(r *http.Request, t *model.Transaction) *model.Transaction {
	k := datastore.NewIncompleteKey(d.ctx, "Transaction", nil)
	t.InsertMoment = time.Now().UnixNano()
	k, err := datastore.Put(d.ctx, k, t)
	api.PanicIfNotNil(err)

	t.IdVal = k.IntID()
	k, err = datastore.Put(d.ctx, k, t)
	api.PanicIfNotNil(err)

	return t
}

func (d DatabaseDatastore) GetParentTransaction(r *http.Request, transactionID string) *model.Transaction {
	var arr []model.Transaction
	q := datastore.NewQuery("Transaction").Filter("Sign =", transactionID).Limit(1)
	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)

	if len(arr) == 0 {
		api.PanicWithMsg("Invalid parent")
	}

	return &arr[0]
}

/*
	Contracts
*/
func (d DatabaseDatastore) InsertSignRequest(r *http.Request, t *model.Transaction) *model.Transaction {
	k := datastore.NewIncompleteKey(d.ctx, "SignRequest", nil)
	t.Creation = time.Now().UnixNano()
	k, err := datastore.Put(d.ctx, k, t)
	if err != nil {
		panic(err)
	}
	t.IdVal = k.IntID()
	k, err = datastore.Put(d.ctx, k, t)
	if err != nil {
		panic(err)
	}
	return t
}

func (d DatabaseDatastore) GetSignRequest(r *http.Request, id int64) *model.Transaction {
	k := datastore.NewKey(d.ctx, "SignRequest", "", id, nil)
	t := new(model.Transaction)
	err := datastore.Get(d.ctx, k, t)
	api.PanicIfNotNil(err)
	return t
}

/*
	Blocks
*/
func (d DatabaseDatastore) InsertBlockRequest(blreq *model.BlockRequest) {
	k := datastore.NewIncompleteKey(d.ctx, "BlockRequest", nil)
	k, err := datastore.Put(d.ctx, k, blreq)
	api.PanicIfNotNil(err)

	blreq.IdVal = k.IntID()
	_, err = datastore.Put(d.ctx, k, blreq)
	api.PanicIfNotNil(err)
}

func (d DatabaseDatastore) InsertBlockRequestForInstance(transport *model.BlockRequestForInstance) {
	k := datastore.NewIncompleteKey(d.ctx, "BlockRequestForInstance", nil)
	k, err := datastore.Put(d.ctx, k, transport)
	api.PanicIfNotNil(err)

	transport.NodeIdVal = k.IntID()
	k, err = datastore.Put(d.ctx, k, transport)
	api.PanicIfNotNil(err)
}
func (d DatabaseDatastore) UpdateBlockRequestForInstance(transport *model.BlockRequestForInstance) {
	k := datastore.NewKey(d.ctx, "BlockRequestForInstance", "", transport.NodeIdVal, nil)
	k, err := datastore.Put(d.ctx, k, transport)
	api.PanicIfNotNil(err)
}
func (d DatabaseDatastore) GetBlockRequestForInstanceFromHash(hash string) *model.BlockRequestForInstance {
	q := datastore.NewQuery("BlockRequestForInstance").Filter("BlockSignProposal = ", hash).Limit(1)
	count, err := q.Count(d.ctx)
	api.PanicIfNotNil(err)

	if count == 0 {
		api.PanicWithMsg("Invalid Block")
	}

	brfi := &model.BlockRequestForInstance{}
	ti := q.Run(d.ctx)
	_, err = ti.Next(brfi)
	api.PanicIfNotNil(err)

	return brfi
}

func (d DatabaseDatastore) InsertBlock(block *model.Block) {
	k := datastore.NewIncompleteKey(d.ctx, "Block", nil)
	_, err := datastore.Put(d.ctx, k, block)
	api.PanicIfNotNil(err)
}
func (d DatabaseDatastore) CountBlocks() int {
	q := datastore.NewQuery("Block")
	count, err := q.Count(d.ctx)
	api.PanicIfNotNil(err)
	return count
}
func (d DatabaseDatastore) GetLastBlocks(size int) []model.Block {
	var arr []model.Block
	q := datastore.NewQuery("Block").Order("-Creation").Limit(size)

	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)

	return arr
}
func (d DatabaseDatastore) BlockTransactionsCursor(sign string) model.StorageCursor {
	//todo: fixed to gae

	var ret model.StorageCursor
	q := datastore.NewQuery("Transaction").Project("InsertMoment", "Sign").Filter("BlockSign = ", sign).Order("InsertMoment")
	ret.GAE = q.Run(d.ctx)

	return ret
}
func (d DatabaseDatastore) NextTransactionSign(cursor model.StorageCursor) ([]byte, bool) {
	var p model.Transaction
	_, err := cursor.GAE.Next(&p)
	if err == datastore.Done {
		return nil, true
	}
	if err != nil {
		api.PanicIfNotNil(err)
	}

	log.Infof(d.ctx, "sign: %s, insertmoment: %v", p.Sign, p.InsertMoment)

	sign := crypto.Base64_decode(p.Sign)
	return sign, false
}

func (d DatabaseDatastore) BlockTransactionsCursorClose(mod model.StorageCursor) {

}

/*
Query
*/

func (d DatabaseDatastore) GetBlocksByOffset(size, offset int) []model.Block {
	if size > 100 {
		size = 100
	}
	var arr []model.Block
	q := datastore.NewQuery("Block").Order("-Creation").Offset(offset).Limit(size)

	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)

	return arr
}

func (d DatabaseDatastore) GetBlockTransactions(blockid string, size, offset int, metadata bool) []model.Transaction {
	panic("implement me")
}

func (d DatabaseDatastore) GetGroupTransactions(blockid string, size, offset int, metadata bool) []model.Transaction {
	panic("implement me")
}

/*
	Query Explorer
*/
func (d DatabaseDatastore) GetApplicationTransactions(appid, from, to string, metadata bool) []model.Transaction {
	q := datastore.NewQuery("Transaction")

	if metadata {
		q = q.Project("Sign", "Creation", "Hash", "Parent", "AppID", "SignerKinds", "SignKind")
	}

	log.Infof(d.ctx, appid)

	q = q.Filter("AppID = ", appid)

	if from != "" {
		f, err := strconv.ParseInt(from, 10, 64)
		api.PanicIfNotNil(err)

		q = q.Filter("InsertMoment > ", f)
		log.Infof(d.ctx, "from: %d", f)
	}

	if to != "" {
		t, err := strconv.ParseInt(to, 10, 64)
		api.PanicIfNotNil(err)

		q = q.Filter("InsertMoment < ", t)
		log.Infof(d.ctx, "to: %d", t)
	}

	log.Infof(d.ctx, "query: %+v", q)

	q = q.Order("InsertMoment").Limit(100)

	var arr []model.Transaction
	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)

	return arr
}

func (d DatabaseDatastore) GetApplicationTransaction(appid, sign string, metadata bool) *model.Transaction {
	q := datastore.NewQuery("Transaction")

	if metadata {
		q = q.Project("Sign", "Creation", "Hash", "Parent", "AppID", "SignerKinds", "SignKind")
	}

	q = q.Filter("AppID = ", appid)

	q = q.Filter("Sign = ", sign).Limit(1)

	var arr []model.Transaction
	_, err := q.GetAll(d.ctx, &arr)
	api.PanicIfNotNil(err)

	if len(arr) == 0 {
		return nil
	}
	return &arr[0]
}

func (d DatabaseDatastore) GetApplicationGroupTransactions(appid, groupsign string, metadata bool) []model.Transaction {
	q := datastore.NewQuery("Transaction")

	if metadata {
		q = q.Project("Sign", "Creation", "Hash", "Parent", "AppID", "SignerKinds", "SignKind")
	}

	q = q.Filter("AppID = ", appid).Filter("Parent = ", groupsign).Order("InsertMoment")

	var arr []model.Transaction
	_, err := q.GetAll(d.ctx, &arr)

	if err != nil {
		panic(err)
	}

	return arr
}
