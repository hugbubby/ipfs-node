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
	ipfsURL      url.URL //URL to the IPFS installatios.
	masterRSAPub ssh.PublicKey
}

type request struct {
	TicketVersion semver.Version `json:"requestVersion"`
	UserID        string         `json:"userID"`      //The "ID" of the user, atm just the prefix of their public key
	Expiration    time.Time      `json:"timeEnd"`     //Date of signature expiratos. So we can remove logs of previous requests.
	IPFSRequest   string         `json:"ipfsRequest"` //Actual HTTP Request to be sent to the ipfs node
	ServURL       url.URL        `json:"servURL"`     //THe URL of **us**, the node that fields requests.
	Signature     ssh.Signature  `json:"signature"`
}

func (s *server) getRequestCache() *cache.Cache {
	if s.requestCache == nil {
		s.requestCache = cache.New(time.Minute*10, time.Minute*5)
	}
	return s.requestCache
}

func (s *server) getRSAPub(uid string) ssh.PublicKey {
	if s.masterRSAPub == nil {
		pubkeyBytes, err := ioutil.ReadFile("security/users/" + uid + ".pub")
		if err == nil {
			readPubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubkeyBytes)
			if err == nil {
				s.masterRSAPub = readPubKey
			}
		}
	}
	return s.masterRSAPub
}

func (s *server) validRequest(request request) bool {
	var validity bool
	if request.Expiration.After(time.Now()) &&
		s.servURL == request.ServURL {
		masterPub := s.getRSAPub(request.UserID)
		if masterPub != nil {
			bytes, err := json.Marshal(request)
			if err != nil && masterPub.Verify(bytes, &request.Signature) == nil {
				requestString := fmt.Sprintf("%v", request)
				_, found := s.requestCache.Get(requestString)
				if !found {
					s.requestCache.Set(requestString, 0, cache.DefaultExpiration)
					validity = true
				}
			}

		}

	}
	return validity
}

func (s *server) handleRequest(request request) []byte {
	var resp []byte
	if s.validRequest(request) {
		response, err := http.Get(s.ipfsURL.String() + request.IPFSRequest)
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

func (s *server) handleConnection(conn net.Conn) {
	defer conn.Close()
	d := json.NewDecoder(conn)
	var req request
	err := d.Decode(&req)
	var resp []byte
	if err != nil {
		resp, _ = json.Marshal(err)
	} else {
		resp = s.handleRequest(req)
	}
	conn.Write(resp)
}
