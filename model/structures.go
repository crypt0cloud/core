package model

import "net/http"

type ModelConnector interface {
	Open(r *http.Request, config string) (ModelDatabase)
	/**
	Coordinator
	 */


}

type ModelDatabase interface {
	Coord_MasterKeyExists() bool
	Coord_InsertKey(key *MasterKey)
}

/**
	Structures for Coordinator
 */

type MasterKey struct {
	MasterPublicKey []byte
	URL	string
	CoordinatorPublic	[]byte
	CoordinatorPrivate	[]byte
}

type ChainNode struct {
	URL	string
}