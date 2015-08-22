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
	minion := RemoteMinion{client}

	log.Printf("Got a connection from %s\n", conn.CommonNames[0])

	arches, err := minion.Arches()
	if err != nil {
		log.Fatalf("Ouch: %s\n", err)
	}

	log.Printf(" -> They can do %s", strings.Join(arches, ", "))

	ftbfs, err := minion.Build(
		[]Archive{Archive{}},
		ChrootTarget{Chroot: "unstable", Suite: "unstable"},
		"amd64", "pool/f/fbautostart_fnord.dsc",
	)
	log.Printf("Heard back: %s %s\n", ftbfs, err)
}
