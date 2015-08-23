package main

import (
	"fmt"
	"log"
	"net/rpc"

	"pault.ag/go/service"
)

/* */

func BeAMinion(cert, key, ca, host string, port int) {
	node := MinionNode{}
	log.Printf("Bringing Minion online\n")
	node.Register()
	log.Printf("Diling coordinator\n")
	conn, err := service.DialFromKeys(
		fmt.Sprintf("%s:%d", host, port),
		cert, key, ca,
	)
	if err != nil {
		log.Fatalf("Error! %s\n", err)
	}
	log.Printf("Doing what they say!\n")
	rpc.ServeConn(conn)
}

/* */

type MinionNode struct{ service.Node }

func (m *MinionNode) Register() {
	minion := Minion{Arches: []string{"amd64", "all"}}
	rpc.Register(&minion)
}
