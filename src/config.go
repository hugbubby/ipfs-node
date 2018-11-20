package main

import "encoding/json"
import "os"
import "net/url"

//Main configuration structure.
//Unmarshal json config into one of these
//structs.
type configuration struct {
	ServerURL string `json:"serverURL"`
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

func (config *configuration) makeServer() server {
	var s server

	ipfsURL, err := url.Parse(config.IpfsURL)
	if err == nil {
		s.ipfsURL = ipfsURL
	}

	serverURL, err := url.Parse(config.ServerURL)
	if err == nil {
		s.URL = serverURL
	}

	return s
}
