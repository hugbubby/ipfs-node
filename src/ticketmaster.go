package main

import "github.com/blang/semver"
import ipfs "github.com/ipfs/go-ipfs-api"
import "net"
import "net/url"
import "time"
import "errors"
import "encoding/json"
import "io/ioutil"
import "golang.org/x/crypto/ssh"

type ticketmaster struct {
	nodeVersion  semver.Version
	shell        *ipfs.Shell
	servURL      url.URL
	ipfsURL      url.URL //URL to the IPFS installation.
	masterRSAPub ssh.PublicKey
}

type pinTicket struct {
	TicketVersion semver.Version `json:"ticketVersion"`
	//TODO: add ticket id so that replay attacks are not possible
	Expiration time.Time     `json:"timeEnd"`    //Date of ticket expiraton. So we can remove logs of previous tickets.
	ObjectHash string        `json:"objectHash"` //IPFS Object to be pinned.
	ServURL    url.URL       `json:"servURL"`    //THe URL of the node that pins
	Signature  ssh.Signature `json:"signature"`
}

func (n *ticketmaster) getMasterRSAPub() ssh.PublicKey {
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
func (n *ticketmaster) validPinTicket(ticket pinTicket) bool {
	var validity bool
	bytes, err := json.Marshal(ticket)
	if err != nil {
		masterPub := n.getMasterRSAPub()
		if masterPub != nil {
			if masterPub.Verify(bytes, &ticket.Signature) == nil {
				validity = (n.servURL == ticket.ServURL)
			}
		}
	}
	return validity
}

func (n *ticketmaster) handlePinRequest(ticket pinTicket) error {
	var err error
	if n.validPinTicket(ticket) {
		if ticket.Expiration.After(time.Now()) {
			shell := n.getIPFSShell()
			err = shell.Pin(ticket.ObjectHash)
		} else {
			err = errors.New("ticket has expired")
		}
	} else {
		err = errors.New("invalid authentication")
	}
	return err
}

func (n *ticketmaster) handleConnection(conn net.Conn) {
	defer conn.Close()
	d := json.NewDecoder(conn)
	var ticket pinTicket
	err := d.Decode(&ticket)
	var resp []byte
	if err != nil {
		resp, _ = json.Marshal(err)
	} else {
		resp, _ = json.Marshal(n.handlePinRequest(ticket))
	}
	conn.Write(resp)
}

//Retreives an IPFS shell or creates one.
func (n *ticketmaster) getIPFSShell() *ipfs.Shell {
	if n.shell == nil || !n.shell.IsUp() {
		if n.ipfsURL.String() == "" {
			n.shell = ipfs.NewLocalShell()
		} else {
			n.shell = ipfs.NewShell(n.ipfsURL.String())
		}
	}
	return n.shell
}
