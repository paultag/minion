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
	log.Printf("Queueing build\n")
	proxy.QueueBuild(minion.Build{
		Chroot: minion.Chroot{
			Chroot: "unstable",
			Target: "unstable",
		},
		Arch: "amd64",
		DSC:  "https://people.debian.org/~paultag/tmp/fluxbox_1.3.6~rc1-1.dsc",
	})
	log.Printf("Queued\n")
}
