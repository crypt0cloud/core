package model

import "net/http"

type ModelConnector interface {
	Open(r *http.Request, config string) ModelDatabase
	/**
	Coordinator
	*/

}

type ModelDatabase interface {
	Coord_MasterKeyExists() bool
	Coord_InsertKey(key *MasterKey)
	Coord_GetKey() *MasterKey

	Coord_Insert_ExternalNode(node *NodeIdentification)
	Coord_GetRandomNodeIdentification(limit int) []NodeIdentification

	GetNodeId() *NodeIdentification
	IsRegisteredNodeID() bool
	RegisteredNodeID(us *NodeIdentification)
	SetupSetEndPointIfNull(endpoint string) bool
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
}

type Transaction struct {
	IdVal int64

	//Control
	OriginatorURl string

	//Signed
	InsertMoment int64
	Sign         string
	Signer       string

	Hash    string
	Content string

	FromNode, ToNode NodeIdentification

	Payload     string
	Parent      int64
	ParentBlock int64 //TODO AGREGARIN SINGLE TRANSACTIONS, Y CONTRACT CREATION
	AppID       string
	SignerKinds []string
	SignKind    string
	Callback    string
}
