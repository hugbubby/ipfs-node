package main

import "github.com/blang/semver"
import "net"
import "net/url"
import "net/http"
import "time"
import "errors"
import "encoding/json"
import "io/ioutil"
import "golang.org/x/crypto/ssh"

type server struct {
	nodeVersion  semver.Version
	servURL      url.URL
	ipfsURL      url.URL //URL to the IPFS installation.
	masterRSAPub ssh.PublicKey
}

type request struct {
	TicketVersion semver.Version `json:"requestVersion"`
	//TODO: add request id so that replay attacks are not possible
	Expiration  time.Time     `json:"timeEnd"`     //Date of request expiraton. So we can remove logs of previous requests.
	IPFSRequest string        `json:"ipfsRequest"` //Actual HTTP Request to be sent to the ipfs node
	ServURL     url.URL       `json:"servURL"`     //THe URL of **us**, the node that pins
	Signature   ssh.Signature `json:"signature"`
}

type requestContents struct {
}

func (n *server) getMasterRSAPub() ssh.PublicKey {
	if n.masterRSAPub == nil {
		pubkeyBytes, err := ioutil.ReadFile("security/master.pub")
		if err == nil {
			readPubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubkeyBytes)
			if err == nil {
				n.masterRSAPub = readPubKey
			}
		}
	}
	return n.masterRSAPub
}

//TODO: Determines validity of pinrequest
func (n *server) validRequest(request request) bool {
	var validity bool
	bytes, err := json.Marshal(request)
	masterPub := n.getMasterRSAPub()
	if err != nil &&
		masterPub != nil &&
		masterPub.Verify(bytes, &request.Signature) == nil &&
		request.Expiration.After(time.Now()) &&
		n.servURL == request.ServURL {
		validity = true
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
