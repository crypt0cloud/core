package model

import (
	"net/http"
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
	InsertTransaction(r *http.Request, t *Transaction) *Transaction
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
	Creation   int64
	PublicKey  string
	PrivateKey string `json:",omitempty"`
	Endpoint   string
	Myself     bool
}

type Transaction struct {
	IdVal int64

	//Control
	OriginatorURl string

	//Signed
	InsertMoment int64
	Sign         string
	Signer       string

	Hash     string
	Content  string
	Creation int64

	FromNode, ToNode NodeIdentification

	Payload     string
	Parent      int64
	ParentBlock int64 //TODO AGREGARIN SINGLE TRANSACTIONS, Y CONTRACT CREATION
	AppID       string
	SignerKinds []string
	SignKind    string
	Callback    string
}
