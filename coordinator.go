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

	BuildChannels *minion.BuildChannelMap
}

func (m *coordinatorService) Register() {
	remote := minion.NewCoordinatorRemote(m.BuildChannels)
	rpc.Register(&remote)
}

func (m *coordinatorService) Handle(rpcClient *rpc.Client, conn *service.Conn) {
	log.Printf("Got a connection from %s\n", conn.Name)
	client := minion.MinionProxy{rpcClient}

	arches, err := client.GetArches()
	if err != nil {
		log.Printf("Error: %s\n", err)
		return
	}

	for {
		job := <-m.BuildChannels.Get(arches[0])
		client.Build(job)
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

	buildChannel := minion.BuildChannelMap{}
	coordinator := coordinatorService{
		BuildChannels: &buildChannel,
	}
	log.Printf("Great, waiting for Minions, and telling them what to do!\n")
	service.Handle(l, &coordinator)

}
