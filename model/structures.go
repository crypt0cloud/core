package model

import (
	"google.golang.org/appengine/datastore"
	"net/http"
	"time"
)

type ModelConnector interface {
	Open(r *http.Request, config string) ModelDatabase
	/**
	Coordinator
	*/

}

type ModelDatabase interface {
	/*
		Coordinator
	*/
	Coord_MasterKeyExists() bool
	Coord_InsertKey(key *MasterKey)
	Coord_GetKey() *MasterKey

	Coord_Insert_ExternalNode(node *NodeIdentification)
	Coord_GetRandomNodeIdentification(limit int) []NodeIdentification

	/*
		Nodes
	*/
	GetNodeId() *NodeIdentification
	IsRegisteredNodeID() bool
	RegisteredNodeID(us *NodeIdentification)
	SetupSetEndPointIfNull(endpoint string) bool

	/*
		Apps
	*/
	AppIdExists(r *http.Request, id string) bool

	/*
		Ledger
	*/
	NodeExists(r *http.Request) bool
	GetCoordinatorKey(r *http.Request) *Transaction
	UserSignExist(r *http.Request, id string) *Transaction
	UserPayloadExist(r *http.Request, id string) *Transaction

	/*
		Transactions
	*/
	InsertTransaction(r *http.Request, t *Transaction) *Transaction
	GetParentTransaction(r *http.Request, transactionID string) *Transaction
	/*
		Contracts
	*/
	InsertSignRequest(r *http.Request, t *Transaction) *Transaction
	GetSignRequest(r *http.Request, id int64) *Transaction

	/*
		Blocks
	*/
	InsertBlock(block *Block)
	CountBlocks() int
	GetLastBlocks(size int) []Block

	BlockTransactionsCursor(sign string) StorageCursor
	NextTransactionSign(cursor StorageCursor) ([]byte, bool)
	BlockTransactionsCursorClose(mod StorageCursor)

	/*
		Query
	 */
	GetBlocks(size, offset int) []Block
	GetBlockTransactions(blockid string, size, offset int, metadata bool) []Transaction
	GetGroupTransactions(blockid string, size, offset int, metadata bool) []Transaction

	/*
		Query Explorer
	 */


}

/**
Structures for Coordinator
*/

type MasterKey struct {
	MasterPublicKey    []byte
	URL                string
	CoordinatorPublic  []byte
	CoordinatorPrivate []byte
}

/**
Structures for nodes
*/
type NodeIdentification struct {
	Creation   int64  `json:",omitempty"`
	PublicKey  string `json:",omitempty"`
	PrivateKey string `json:",omitempty"`
	Endpoint   string `json:",omitempty"`
	Myself     bool
}

type Block struct {
	Creation                  time.Time
	TransactionsCount         int
	NextBlockTransactionsUsed int
	Hash                      string
	Sign                      string
	BlockTime                 time.Time
}

type Transaction struct {
	IdVal int64

	BlockSign string `json:",omitempty"`

	//Control
	//OriginatorURl string

	//Signed
	InsertMoment int64  `json:",omitempty"`
	Sign         string `json:",omitempty"`
	Signer       string `json:",omitempty"`

	Hash     string `json:",omitempty"`
	Content  string `datastore:",noindex" json:",omitempty"`
	Creation int64  `json:",omitempty"`

	FromNode NodeIdentification `json:",omitempty"`
	ToNode NodeIdentification `json:",omitempty"`

	Payload string `datastore:",noindex" json:",omitempty"`
	Parent  string `json:",omitempty"`
	//ParentBlock int64 //TODO AGREGARIN SINGLE TRANSACTIONS, Y CONTRACT CREATION
	AppID       string   `json:",omitempty"`
	SignerKinds []string `json:",omitempty"`
	SignKind    string   `json:",omitempty"`
	Callback    string   `json:",omitempty"`
}

type StorageCursor struct {
	GAE *datastore.Iterator
}
