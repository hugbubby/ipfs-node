package main

import "fmt"
import "net"
import "net/url"
import "net/http"
import "time"
import "errors"
import "encoding/json"
import "io/ioutil"
import "golang.org/x/crypto/ssh"
import "github.com/blang/semver"
import "github.com/patrickmn/go-cache"

type server struct {
	requestCache *cache.Cache
	nodeVersion  semver.Version
	servURL      url.URL
	ipfsURL      url.URL //URL to the IPFS installation.
	masterRSAPub ssh.PublicKey
}

type request struct {
	TicketVersion semver.Version `json:"requestVersion"`
	UserID        string         `json:"userID"`
	Expiration    time.Time      `json:"timeEnd"`     //Date of signature expiraton. So we can remove logs of previous requests.
	IPFSRequest   string         `json:"ipfsRequest"` //Actual HTTP Request to be sent to the ipfs node
	ServURL       url.URL        `json:"servURL"`     //THe URL of **us**, the node that fields requests.
	Signature     ssh.Signature  `json:"signature"`
}

func (n *server) getRSAPub(uid string) ssh.PublicKey {
	if n.masterRSAPub == nil {
		pubkeyBytes, err := ioutil.ReadFile("security/users/" + uid + ".pub")
		if err == nil {
			readPubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubkeyBytes)
			if err == nil {
				n.masterRSAPub = readPubKey
			}
		}
	}
	return n.masterRSAPub
}

func (n *server) validRequest(request request) bool {
	var validity bool
	if request.Expiration.After(time.Now()) &&
		n.servURL == request.ServURL {
		masterPub := n.getRSAPub(request.UserID)
		if masterPub != nil {
			bytes, err := json.Marshal(request)
			if err != nil && masterPub.Verify(bytes, &request.Signature) == nil {
				requestString := fmt.Sprintf("%v", request)
				_, found := n.requestCache.Get(requestString)
				if !found {
					n.requestCache.Set(requestString, 0, cache.DefaultExpiration)
					validity = true
				}
			}

		}

	}
	return validity
}

func (n *server) handleRequest(request request) []byte {
	var resp []byte
	if n.validRequest(request) {
		response, err := http.Get(n.ipfsURL.String() + request.IPFSRequest)
		if err != nil {
			resp, _ = json.Marshal(err)
		} else {
			response.Body.Read(resp)
		}
	} else {
		resp, _ = json.Marshal(errors.New("invalid authentication"))
	}
	return resp
}

func (n *server) handleConnection(conn net.Conn) {
	defer conn.Close()
	d := json.NewDecoder(conn)
	var req request
	err := d.Decode(&req)
	var resp []byte
	if err != nil {
		resp, _ = json.Marshal(err)
	} else {
		resp = n.handleRequest(req)
	}
	conn.Write(resp)
}
