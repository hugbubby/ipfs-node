package main

import "encoding/json"
import "os"

//Main configuration structure.
//Unmarshal json config into one of these
//structs.
type configuration struct {
	ServerAddress string `json:"serverAddress"`
	IpfsAddress   string `json:"ipfsAddress"`
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

	s.ipfsAddress = config.IpfsAddress
	s.address = config.ServerAddress

	return s
}
