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

var suites *string

func init() {
	suites = minionCommand.Flag.String("suites", "", "comma seperated suite:arch pairs")
}

type minionService struct {
	service.Node

	Config minion.MinionConfig
}

func (m *minionService) Register() {
	if *suites == "" {
		log.Fatalf("No suites given\n")
	}

	buildableSuites := []minion.BuildableSuite{}
	suitePairs := strings.Split(*suites, ",")
	for _, suitePair := range suitePairs {
		pair := strings.Split(suitePair, ":")
		if len(pair) != 2 {
			panic(fmt.Errorf("Error! %s is an invalid suite pair", suitePair))
		}

		buildableSuites = append(buildableSuites, minion.BuildableSuite{
			Suite: pair[0],
			Arch:  pair[1],
		})
	}

	minion := minion.NewMinionRemote(m.Config, buildableSuites)
	rpc.Register(&minion)
}

func minionRun(config minion.MinionConfig, cmd *Command, args []string) {
	log.Printf("Bringing Minion online\n")
	node := minionService{Config: config}
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
