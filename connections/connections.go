package connections

import (
	"bytes"
	"encoding/json"
	md "github.com/crypt0cloud/core/model"
	gae "github.com/crypt0cloud/model_gae"
	"log"

	"github.com/onlyangel/apihandlers"
	"io/ioutil"
	"net/http"
)

var model md.ModelConnector

func init() {
	var err error
	model, err = md.Open("datastore")
	if err != nil {
		log.Fatal(err)
	}
}

func GetRemoteNodeCredentials(r *http.Request, endpoint string) *md.NodeIdentification {
	//TODO: CHANGE URL WHEN BLOCK CHANGES
	response, err := CallRemote(r, "http://"+endpoint+"/api/v1/node_id")
	apihandlers.PanicIfNotNil(err)

	nodeI := new(md.NodeIdentification)
	err = json.Unmarshal(response, nodeI)
	apihandlers.PanicIfNotNil(err)

	return nodeI
}

func CallRemote(r *http.Request, url string) ([]byte, error) {
	// Todo: Replace gae.GetClient
	client := gae.GetClient(r)
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func PostRemote(r *http.Request, url string, data []byte) []byte {
	log.Printf("url:")
	log.Printf(url)

	client := gae.GetClient(r)

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	apihandlers.PanicIfNotNil(err)

	body, err := ioutil.ReadAll(resp.Body)
	apihandlers.PanicIfNotNil(err)

	defer resp.Body.Close()

	return body
}

func ValidateTransactionWithPeers(r *http.Request, transaction *md.Transaction, fromnode md.NodeIdentification) bool {

	jsonstr, err := json.Marshal(transaction)
	apihandlers.PanicIfNotNil(err)

	response := PostRemote(r, "http://"+fromnode.Endpoint+"/api/v1/coord/verify_with_peers", jsonstr)

	return string(response) == "true"
}
