package main

import (
	"log"
	"net/rpc"
	"os"
	"strings"

	"pault.ag/go/service"
)

/* */

func BeAMinion() {
	node := MinionNode{}
	log.Printf("Bringing Minion online\n")
	node.Register()
	log.Printf("Diling coordinator\n")
	conn, err := service.DialFromKeys(
		"cassiel.pault.ag:8888",
		"certs/personal.crt", "certs/personal.key",
		"certs/ca.crt",
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

/* */

func BeACoordinator() {
	log.Printf("Bringing TCP server online!\n")
	l, err := service.ListenFromKeys(
		"cassiel.pault.ag:8888",
		"certs/cassiel.crt", "certs/cassiel.key",
		"certs/ca.crt",
	)
	if err != nil {
		log.Fatalf("Server Ouchie! %s", err)
	}
	coordinator := MinionCoordinator{}
	log.Printf("Great, waiting for Minions, and telling them what to do!\n")
	service.Handle(l, &coordinator)
}

/**/

type MinionCoordinator struct{ service.Coordinator }

func (m *MinionCoordinator) Handle(client *rpc.Client, conn *service.Conn) {
	log.Printf("Got a connection from %s\n", conn.CommonNames[0])
	var args interface{}
	arches := []string{}
	if err := client.Call("Minion.GetArches", &args, &arches); err != nil {
		log.Fatalf("Error!: %s\n", err)
	}
	log.Printf(" -> They can do %s", strings.Join(arches, ", "))
}

func main() {
	switch os.Args[1] {
	case "minion":
		BeAMinion()
	case "coordinator":
		BeACoordinator()
	default:
		log.Fatalf("Don't know what to do :(\n")
	}
}
