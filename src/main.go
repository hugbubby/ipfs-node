package main

import "log"

func main() {
	log.SetFlags(log.Lshortfile)

	config := new(configuration)
	config.loadFile("config.json")
	server := config.makeServer()
	server.start()
}
