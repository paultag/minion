package main

import (
	"fmt"
	"log"
	"net/rpc"

	"pault.ag/go/minion/minion"
	"pault.ag/go/service"
)

var minionCommand = Command{
	Name:  "minion",
	Run:   minionRun,
	Usage: ``,
}

type minionService struct {
	service.Node
}

func (m *minionService) Register() {
	minion := minion.MinionRemote{Arches: []string{"amd64", "all"}}
	rpc.Register(&minion)
}

func minionRun(config MinionConfig, cmd *Command, args []string) {
	log.Printf("Bringing Minion online\n")
	node := minionService{}
	node.Register()
	log.Printf("Diling coordinator\n")
	conn, err := service.DialFromKeys(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		config.Cert, config.Key, config.CaCert,
	)
	if err != nil {
		log.Fatalf("Error! %s\n", err)
	}
	log.Printf("Doing what they say!\n")
	service.ServeConn(conn)
}
