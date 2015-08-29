package main

import (
	"fmt"
	"log"

	"pault.ag/go/minion/minion"
	"pault.ag/go/service"
)

var remoteCommand = Command{
	Name:  "remote",
	Run:   remoteRun,
	Usage: ``,
}

func remoteRun(config MinionConfig, cmd *Command, args []string) {
	conn, err := service.DialFromKeys(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		config.Cert, config.Key, config.CaCert,
	)
	if err != nil {
		log.Fatalf("Error! %s\n", err)
	}
	proxy := minion.CoordinatorProxy{service.Client(conn)}
	proxy.QueueBuild(minion.Build{
		Arch: "amd64",
		DSC:  "https://something/f/fnord.dsc",
	})
}
