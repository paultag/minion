package main

import (
	"fmt"
	"log"
	"net/rpc"
	"pault.ag/go/service"
	"strings"
)

/* */

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

/* */

func BeACoordinator(cert, key, ca, host string, port int) {
	log.Printf("Bringing TCP server online!\n")
	l, err := service.ListenFromKeys(
		fmt.Sprintf("%s:%d", host, port),
		cert, key, ca,
	)
	if err != nil {
		log.Fatalf("Server Ouchie! %s", err)
	}
	coordinator := MinionCoordinator{}
	log.Printf("Great, waiting for Minions, and telling them what to do!\n")
	service.Handle(l, &coordinator)
}

/**/
