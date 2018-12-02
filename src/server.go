package main

import "fmt"
import "log"
import "io"
import "io/ioutil"
import "net/http"
import "net/http/httputil"
import "crypto/x509"
import "encoding/pem"
import "github.com/dgrijalva/jwt-go"

type server struct {
	address      string //URL of server.
	ipfsAddress  string
	tlsCertPath  string
	tlsKeyPath   string
	tokenKey     interface{} //RSA, ECDSA, ETC
	tokenKeyPath string
}

func (s *server) getTokenKeyPath() string {
	if s.tokenKeyPath == "" {
		s.tokenKeyPath = "security/tokens/privkey.pem"
	}
	return s.tokenKeyPath
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
	var err error
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		err = fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	} else {
		if s.tokenKey == nil {
			var bytes []byte
			bytes, err = ioutil.ReadFile(s.getTokenKeyPath())
			if err == nil {
				block, _ := pem.Decode(bytes)
				if block != nil {
					var key interface{}
					key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
					if err == nil {
						s.tokenKey = key
					}
				} else {
					err = fmt.Errorf("failed to decode token authentication key - is it in the PEM format?")
				}
			}
		}
	}
	return s.tokenKey, err
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
