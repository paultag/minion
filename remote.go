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

	if len(args) == 0 {
		log.Fatalf("No subcommand given")
	}

	switch args[0] {
	case "backfill":
		Backfill(config, proxy, args[1:])
	case "status":
		Status(config, proxy, args[1:])
	case "binNMU":
		BinNMU(config, proxy, args[1:])
	}
}

func Status(config minion.MinionConfig, proxy minion.CoordinatorProxy, args []string) {
	queueLengths, err := proxy.GetQueueLengths()
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	for name, length := range queueLengths {
		fmt.Printf("%s - %d pending job(s)\n", name, length)
	}
}

func BinNMU(config minion.MinionConfig, proxy minion.CoordinatorProxy, args []string) {
	log.Fatalf("Unimplemented")
}

func Backfill(config minion.MinionConfig, proxy minion.CoordinatorProxy, args []string) {
	for _, archive := range args {
		needs, err := proxy.GetBuildNeeding(archive, "unstable", "any", "")
		if err != nil {
			log.Fatalf("%s", err)
		}
		for _, need := range needs {
			log.Printf("%s [%s] - %s", archive, need.Arch, need.Location)
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
