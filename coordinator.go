package main

import (
	"fmt"
	"log"
	"net/rpc"

	"pault.ag/go/minion/minion"
	"pault.ag/go/service"
)

type coordinatorService struct {
	service.Coordinator

	BuildChannels *map[string]chan minion.Build
}

func (m *coordinatorService) Register() {
	remote := minion.NewCoordinatorRemote(m.BuildChannels)
	rpc.Register(&remote)
}

func (m *coordinatorService) Handle(client *rpc.Client, conn *service.Conn) {
	log.Printf("Got a connection from %s\n", conn.Name)
	minion := minion.MinionProxy{client}

	arches, err := minion.GetArches()
	if err != nil {
		log.Printf("Error: %s\n", err)
		return
	}
	log.Printf("%s\n", arches)

	for {
		log.Printf("Consuming\n")
		job := <-(*m.BuildChannels)["amd64"]
		log.Printf("Consumed.\n")
		minion.Build(job)
	}
}

var coordinatorCommand = Command{
	Name:  "coordinator",
	Run:   coordinatorRun,
	Usage: ``,
}

func coordinatorRun(config MinionConfig, cmd *Command, args []string) {
	log.Printf("Bringing coordinator online\n")

	l, err := service.ListenFromKeys(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		config.Cert, config.Key, config.CaCert,
	)
	if err != nil {
		log.Fatalf("Server Ouchie! %s", err)
	}

	buildChannel := map[string]chan minion.Build{}
	coordinator := coordinatorService{
		BuildChannels: &buildChannel,
	}
	log.Printf("Great, waiting for Minions, and telling them what to do!\n")
	service.Handle(l, &coordinator)

}
