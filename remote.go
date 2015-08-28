package main

import (
	"fmt"
	"log"

	"pault.ag/go/service"
)

func BeARemote(config MinionConfig) {
	node := MinionNode{}
	log.Printf("Bringing remote online\n")
	node.Register()
	log.Printf("Diling coordinator\n")
	conn, err := service.DialFromKeys(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		config.Cert, config.Key, config.CaCert,
	)
	if err != nil {
		log.Fatalf("Error! %s\n", err)
	}
	log.Printf("Bringing RPC online")
	client := service.Client(conn)
	log.Printf("%s\n", client)
}

// vim: foldmethod=marker
