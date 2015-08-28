package main

import (
	"fmt"
	"log"
	"net/rpc"
	"strings"

	"pault.ag/go/service"
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
	log.Printf("Writing %s <- %s", job.Arch, job)
	m.Jobs[job.Arch] <- job
}

func (m *MinionCoordinator) Register() {
	log.Printf("Register\n")
}

func (m *MinionCoordinator) Handle(client *rpc.Client, conn *service.Conn) {
	minion := RemoteMinion{client}
	name := conn.CommonNames[0]
	log.Printf("Got a connection from %s\n", name)

	arches, err := minion.Arches()
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}

	log.Printf("%s can handle arches %s\n", name, strings.Join(arches, ", "))

	pipes := []<-chan Job{}
	for _, arch := range arches {
		pipes = append(pipes, m.Jobs[arch])
	}

	for {
		log.Printf("Waiting for a job for %s\n", name)
		job := YakYakYakGetAJob(10, pipes...)
		log.Printf("Got a job!: %s\n", job)
	}
}

/* */

func BeACoordinator(config MinionConfig) {
	log.Printf("Bringing TCP server online!\n")
	l, err := service.ListenFromKeys(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		config.Cert, config.Key, config.CaCert,
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
		DSC:      "pool/f/fnords.dsc",
	})

	log.Printf("Great, waiting for Minions, and telling them what to do!\n")
	service.Handle(l, &coordinator)
}

/**/

// vim: foldmethod=marker
