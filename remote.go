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

func remoteRun(config minion.MinionConfig, cmd *Command, args []string) {
	conn, err := service.DialFromKeys(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		config.Cert, config.Key, config.CaCert,
	)
	if err != nil {
		log.Fatalf("Error! %s\n", err)
	}
	proxy := minion.CoordinatorProxy{service.Client(conn)}

	for _, archive := range args {
		needs, err := proxy.GetBuildNeeding(archive, "unstable", "any", "")
		if err != nil {
			log.Fatalf("%s", err)
		}
		for _, need := range needs {
			log.Printf("Marking %s for build on %s", need.Location, need.Arch)
			QueueBuildNeeding(
				proxy,
				fmt.Sprintf("http://%s/%s", config.Host, archive),
				need,
				"unstable",
				"main",
				config.Host,
				archive,
			)
		}
	}
}
