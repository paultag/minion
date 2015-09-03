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
	log.Printf("Queueing build\n")
	proxy.QueueBuild(minion.Build{
		Archives: []minion.Archive{
			minion.Archive{
				Root:     "http://http.debian.net/debian/",
				Suite:    "experimental",
				Sections: []string{"main"},
			},
		},
		Chroot: minion.Chroot{
			Chroot: "unstable",
			Target: "unstable",
		},
		Arch: "amd64",
		DSC:  "http://http.debian.net/debian/pool/main/f/fbautostart/fbautostart_2.718281828-1.dsc",
		Upload: minion.Upload{
			Host:    "localhost",
			Port:    1984,
			Archive: "foo",
		},
	})
	log.Printf("Queued\n")
}
