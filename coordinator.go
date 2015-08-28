package main

import (
	"fmt"
	"log"
	"net/rpc"
	"pault.ag/go/service"
	"strings"
)

/* */

type Job struct {
	Archives []Archive
	Target   ChrootTarget
	Arch     string
	DSC      string
}

type MinionCoordinator struct {
	service.Coordinator

	Jobs map[string]chan Job
}

func (m *MinionCoordinator) AddJob(job Job) {
	m.Jobs[job.Arch] <- job
}

func (m *MinionCoordinator) Register() {
	log.Printf("Register\n")
}

func (m *MinionCoordinator) Handle(client *rpc.Client, conn *service.Conn) {
	minion := RemoteMinion{client}
	log.Printf("Got a connection from %s\n", conn.CommonNames[0])

	arches, err := minion.Arches()
	if err != nil {
		log.Fatalf("Ouch: %s\n", err)
	}
	log.Printf(" -> They can do %s", strings.Join(arches, ", "))
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
	coordinator := MinionCoordinator{
		Jobs: map[string](chan Job){},
	}

	go coordinator.AddJob(Job{
		Archives: []Archive{},
		Target:   ChrootTarget{Chroot: "unstable", Suite: "unstable"},
		Arch:     "amd64",
		DSC:      "pool/f/fnord.dsc",
	})

	log.Printf("Great, waiting for Minions, and telling them what to do!\n")
	service.Handle(l, &coordinator)
}

/**/

// vim: foldmethod=marker
