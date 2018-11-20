package main

import "encoding/json"
import "strings"
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
	if err == nil {
		var filedirname string
		if filename != "" {
			i := strings.LastIndex(filename, "\\")
			if i == -1 {
				i = strings.LastIndex(filename, "/")
				if i == -1 {
					i = 0
				}
			}
			filedirname = filename[0:i]
		} else {
			filedirname = ""
		}
		filedir, err := os.Open(filedirname)
		if err == nil {
			err = filedir.Chdir()
			if err == nil {
				decoder := json.NewDecoder(file)
				err = decoder.Decode(config)
			}
		}
		filedir.Close()
	}
	file.Close()
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
		s.url = serverURL
	}

	return s
}
