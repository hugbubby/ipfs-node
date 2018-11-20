package main

import "log"

func main() {
	log.SetFlags(log.Lshortfile)

	config := new(configuration)
	err := config.loadFile("config.json")
	if err == nil {
		server := config.makeServer()
		server.start()
	} else {
		log.Fatal(err)
	}
}
