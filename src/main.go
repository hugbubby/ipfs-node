package main

import "log"
import "net/url"
import "crypto/tls"

func main() {
	log.SetFlags(log.Lshortfile)

	config := new(configuration)
	config.loadFile("config.json")
	server := config.Server.toServer()

	cert, err := tls.LoadX509KeyPair("security/server.cert", "security/server.key")
	if err != nil {
		log.Fatal(err)
		return
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	listenURL, err := url.Parse(config.Server.ListenURL)
	if err != nil {
		log.Println("Error parsing configuration. Using defualt url.")
		listenURL, _ = url.Parse("tcp://127.0.0.1:25566")
	}
	log.Println("Listener URL: " + listenURL.String())
	log.Println("Starting listener at port " + listenURL.Port())
	listener, err := tls.Listen("tcp", listenURL.Hostname()+":"+listenURL.Port(), tlsConfig)
	if err != nil {
		log.Fatal(err)
	} else {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			go server.handleConnection(conn)
		}
	}
}
