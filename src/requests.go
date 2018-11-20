package main

import "crypto/tls"
import "fmt"
import "log"
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
	Version      semver.Version
	URL          *url.URL
	ipfsURL      *url.URL //URL to the IPFS installatios.
	userKeys     map[string]ssh.PublicKey
	tlsCert      tls.Certificate
}

type request struct {
	IPFSRequest string `json:"ipfsRequest"` //Actual HTTP Request to be sent to the ipfs node
	Stamp       stamp  `json:"stamp"`       //Everything we need to verify this is an authenticated request
}

type stamp struct {
	UserID     string        `json:"userID"`    //The "ID" of the user, atm just the prefix of their public key
	Signature  ssh.Signature `json:"signature"` //Signature from the user's private key
	Expiration time.Time     `json:"timeEnd"`   //Date of signature expiratos. So we can remove logs of previous requests.
	ServURL    url.URL       `json:"servURL"`   //THe URL of **us**, the node that fields requests. To prevent replay attacks across different nodes.
}

func (s *server) getURL() *url.URL {
	if s.URL == nil {
		log.Println("No server URL configured. Listening on default port 25566.")
		s.URL, _ = url.Parse("tcp://127.0.0.1:25566")
	}
	return s.URL
}

func (s *server) getRequestCache() *cache.Cache {
	if s.requestCache == nil {
		s.requestCache = cache.New(time.Minute*10, time.Minute*5)
	}
	return s.requestCache
}

func (s *server) getRSAPub(uid string) ssh.PublicKey {
	key, exists := s.userKeys[uid]
	if !exists {
		pubkeyBytes, err := ioutil.ReadFile("security/users/" + uid + ".pub")
		if err == nil {
			key, _, _, _, err = ssh.ParseAuthorizedKey(pubkeyBytes)
			if err == nil {
				s.userKeys[uid] = key
			}
		}
	}
	return key
}

func (s *server) validRequest(request request) bool {
	var validity bool
	stamp := request.Stamp
	if stamp.Expiration.After(time.Now()) &&
		*s.URL == stamp.ServURL {
		masterPub := s.getRSAPub(stamp.UserID)
		if masterPub != nil {
			bytes, err := json.Marshal(request)
			if err != nil && masterPub.Verify(bytes, &stamp.Signature) == nil {
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
		if s.validRequest(req) {
			response, err := http.Get(s.ipfsURL.String() + req.IPFSRequest)
			if err != nil {
				resp, _ = json.Marshal(err)
			} else {
				response.Body.Read(resp)
			}
		} else {
			resp, _ = json.Marshal(errors.New("invalid authentication"))
		}
	}
	conn.Write(resp)
}

func (s *server) start() {
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{s.tlsCert}}

	servURL := s.getURL()

	log.Println("Listener URL: " + servURL.String())
	log.Println("Starting listener at port " + servURL.Port())
	listener, err := tls.Listen("tcp", servURL.Hostname()+":"+servURL.Port(), tlsConfig)
	if err != nil {
		log.Fatal(err)
	} else {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			go s.handleConnection(conn)
		}
	}
}
