package main

import "encoding/json"
import "os"

//Main configuration structure.
//Unmarshal json config into one of these
//structs.
type configuration struct {
	Server  serverConfig  `json:"server"`
	IpfsAPI ipfsAPIConfig `json:"ipfsAPI"`
}

type serverConfig struct {
	Address      string `json:"address"`
	TLSCertPath  string `json:"tlsCertPath"`
	TLSKeyPath   string `json:"tlsKeyPath"`
	TokenKeyPath string `json:"tokenKeyPath"`
}

type ipfsAPIConfig struct {
	Address string `json:"address"`
}

func (config *configuration) loadFile(filename string) error {
	var err error
	file, err := os.Open(filename)
	if err == nil {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(config)
	}
	file.Close()
	return err
}

func (config *configuration) makeServer() server {
	var s server

	s.address = config.Server.Address
	s.ipfsAddress = config.IpfsAPI.Address

	s.tlsKeyPath = config.Server.TLSKeyPath
	s.tlsCertPath = config.Server.TLSCertPath
	s.tokenKeyPath = config.Server.TokenKeyPath

	return s
}
