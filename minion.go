package main

import (
	"fmt"
	"log"
	"net/rpc"
	"strings"

	"pault.ag/go/minion/minion"
	"pault.ag/go/service"
)

var minionCommand = Command{
	Name:  "minion",
	Run:   minionRun,
	Usage: ``,
}

var archs *string

func init() {
	archs = minionCommand.Flag.String("arch", "", "comma seperated arches")
}

type minionService struct {
	service.Node
}

func (m *minionService) Register() {
	if *archs == "" {
		log.Fatalf("No archs given\n")
	}
	minion := minion.NewMinionRemote(strings.Split(*archs, ","))
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
