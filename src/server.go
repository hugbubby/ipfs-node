package main

import "log"
import "io"
import "io/ioutil"
import "net/http"
import "net/http/httputil"
import "crypto/x509"
import "encoding/pem"
import "github.com/dgrijalva/jwt-go"

type server struct {
	address       string //URL of server.
	ipfsAddress   string
	tlsCertPath   string
	tlsKeyPath    string
	tokenKeys     []interface{} //RSA, ECDSA, ETC //Public Key, not Private
	tokenKeyNames []string
	tokenKeyDir   string
}

func (s *server) getTokenDirPath() string {
	if s.tokenKeyDir == "" {
		s.tokenKeyDir = "security/tokens/pub.cert"
	}
	return s.tokenKeyDir
}

func (s *server) getTokenKeyNames() []string {
	if s.tokenKeyNames == nil {
		tokenKeyFiles, err := ioutil.ReadDir(s.getTokenDirPath())
		if err != nil {
			s.tokenKeyNames = make([]string, len(tokenKeyFiles))
			for i, k := range tokenKeyFiles {
				s.tokenKeyNames[i] = k.Name()
			}
		} else {
			log.Println("error attempting to load keys from", s.getTokenDirPath(), err)
		}
	}
	return s.tokenKeyNames
}

func (s *server) getTokenKeys() ([]interface{}, error) {
	var err error
	if s.tokenKeys == nil {
		s.tokenKeys = make([]interface{}, len(s.getTokenKeyNames()))
		for i, k := range s.getTokenKeyNames() {
			var bytes []byte
			bytes, err = ioutil.ReadFile(s.getTokenDirPath() + k)
			if err == nil {
				block, _ := pem.Decode(bytes)
				if block != nil {
					var key interface{}
					key, err = x509.ParsePKIXPublicKey(block.Bytes)
					if err == nil {
						s.tokenKeys[i] = key
					} else {
						log.Println("failed to parse", k+"'s PEM block - ",
							"is your private key corrupted?!")
					}
				} else {
					log.Println("failed to decode token authentication key - is it in the PEM format?")
				}
			}
		}
	}
	return s.tokenKeys, err
}

func (s *server) getTLSKeyPath() string {
	if s.tlsKeyPath == "" {
		s.tlsKeyPath = "security/tls/privkey.pem"
	}
	return s.tlsKeyPath
}

func (s *server) getTLSCertPath() string {
	if s.tlsCertPath == "" {
		s.tlsCertPath = "security/tls/host.cert"
	}
	return s.tlsCertPath
}

func (s *server) getIPFSAddress() string {
	if s.ipfsAddress == "" {
		s.ipfsAddress = "http://127.0.0.1:5001"
	}
	return s.ipfsAddress
}

func (s *server) getAddress() string {
	if s.address == "" {
		s.address = "http://127.0.0.1:25566"
	}
	return s.address
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (s *server) validToken(token *jwt.Token) bool {
	err := token.Claims.Valid()
	isValid := err != nil
	if isValid {
		log.Println(err)
	}
	return isValid
}

func (s *server) getTokenKey(token *jwt.Token) (interface{}, error) {
	var key interface{}
	keys, err := s.getTokenKeys()
	if err != nil {
		for i, k := range keys {
			if s.getTokenKeyNames()[i] == token.Header["kid"] {
				key = k
				break
			}
		}
	}
	return key, err
}

func (s *server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestDump, err := httputil.DumpRequest(req, false)
	if err == nil {
		log.Printf("New Request:\n%s\n", requestDump)
		if req.Method == "GET" { //IPFS HTTP API only supports get requests.
			tokenS := req.FormValue("ipfsrToken")
			if tokenS != "" {
				t, err := jwt.Parse(tokenS, s.getTokenKey)
				if err == nil {
					if s.validToken(t) {
						var resp *http.Response
						var err error
						if req.FormValue("arg") == "" {
							resp, err = http.Get(s.getIPFSAddress() + req.URL.Path)
						} else {
							resp, err = http.Get(s.getIPFSAddress() + req.URL.Path + "?arg=" + req.FormValue("arg"))
						}
						if err == nil {
							copyHeader(rw.Header(), resp.Header)
							rw.WriteHeader(resp.StatusCode)
							io.Copy(rw, resp.Body)
							resp.Body.Close()
						} else {
							log.Printf("failed to forward authenticated request to IPFS server, %v", err)
						}
					}
				} else {
					log.Printf("error while attempting to parse token from string: %v", err)
				}
			}
		}
	}
}

func (s *server) start() {
	address := s.getAddress()
	log.Println(address)
	err := http.ListenAndServeTLS(address, s.getTLSCertPath(), s.getTLSKeyPath(), s)
	if err != nil {
		log.Fatal(err)
	}
}
