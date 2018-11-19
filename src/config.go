package main

import "github.com/blang/semver"
import "encoding/json"
import "os"
import "net/url"

//Main configuration structure.
//Unmarshal json config into one of these
//structs.
type configuration struct {
	Server serverConfig `json:"ticketmaster"`
}

type serverConfig struct {
	Version   string `json:"version"`
	ListenURL string `json:"listenURL"`
	IpfsURL   string `json:"ipfsURL"`
}

func (config *configuration) loadFile(filename string) error {
	var err error
	file, err := os.Open(filename)
	defer file.Close()
	if err == nil {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(config)
	}
	return err
}

func (cn serverConfig) toServer() server {
	var n server

	version, err := semver.Parse(cn.Version)
	if err == nil {
		n.nodeVersion = version
	}

	ipfsURL, err := url.Parse(cn.IpfsURL)
	if err == nil {
		n.ipfsURL = *ipfsURL
	}

	return n
}
