package main

import (
	"fmt"
	"log"
	"net/rpc"

	"pault.ag/go/mailer"
	"pault.ag/go/minion/minion"
	"pault.ag/go/service"
)

var Mailer *mailer.Mailer

type MailableJob struct {
	Job    minion.Build
	Minion string
}

type coordinatorService struct {
	service.Coordinator

	BuildChannels *minion.BuildChannelMap
	Config        *minion.MinionConfig
}

func (m *coordinatorService) Register() {
	remote := minion.NewCoordinatorRemote(m.BuildChannels, m.Config)
	rpc.Register(&remote)
}

func (m *coordinatorService) Handle(rpcClient *rpc.Client, conn *service.Conn) {
	log.Printf("Got a connection from %s\n", conn.Name)
	client := minion.MinionProxy{rpcClient}

	suites, err := client.GetBuildableSuites()
	if err != nil {
		log.Printf("Error: %s\n", err)
		return
	}

	buildChannels := minion.GetBuildChannels(m.BuildChannels, suites)

	for {
		job := minion.NextBuild(buildChannels, 5)
		log.Printf("Telling %s to build\n", conn.Name)
		ftbfs, err := client.Build(job)
		if err != nil {
			if err == rpc.ErrShutdown {
				log.Printf("Client disconnect: %s - %s\n", conn.Name, err)
				m.BuildChannels.Get(job.GetBuildChannelKey()) <- job
				conn.Close()
				return
			}
			log.Printf("Abnormal exit: %s\n", err)
		}
		if ftbfs {
			log.Printf("FTBFS")

			if Mailer != nil {
				if err := Mailer.Mail(
					[]string{m.Config.Administrator},
					"ftbfs",
					&MailableJob{
						Job:    job,
						Minion: conn.Name,
					},
				); err != nil {
					log.Printf("Error: %s", err)
				}
			}
		}
	}
}

var coordinatorCommand = Command{
	Name:  "coordinator",
	Run:   coordinatorRun,
	Usage: ``,
}

func coordinatorRun(config minion.MinionConfig, cmd *Command, args []string) {
	log.Printf("Bringing coordinator online\n")
	var err error

	if config.Templates != "" {
		Mailer, err = mailer.NewMailer(config.Templates)
		if err != nil {
			log.Fatal("%s\n", err)
		}
	}

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
		Config:        &config,
	}
	log.Printf("Great, waiting for Minions, and telling them what to do!\n")
	service.Handle(l, &coordinator)
}
