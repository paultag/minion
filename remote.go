package main

import (
	"flag"
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
	ensure := func(x *string, arg string) {
		if x != nil && *x != "" {
			return
		}
		log.Fatalf("Missing argument %s", arg)
	}

	flags := flag.FlagSet{}

	dsc := flags.String("dsc", "", "DSC to binNMU")
	archive := flags.String("archive", "", "Archive to binNMU into")
	arch := flags.String("arch", "", "Archive to binNMU into")
	version := flags.String("version", "", "Version to use for the binNMU")
	changes := flags.String("changes", "", "Changes to use for the binNMU")
	suite := flags.String("suite", "", "suite to use for the binNMU")

	flags.Parse(args)

	for _, s := range []struct {
		Name  string
		Value *string
	}{
		{"dsc", dsc},
		{"arch", arch},
		{"archive", archive},
		{"version", version},
		{"changes", changes},
		{"suite", suite},
	} {
		ensure(s.Value, s.Name)
	}

	build := minion.NewBuild(
		config.Host,
		*archive,
		*suite,
		"main",
		*arch,
		*dsc,
	)
	build.BinNMU = minion.BinNMU{
		Version:   *version,
		Changelog: *changes,
	}
	proxy.QueueBuild(build)
}

func Backfill(config minion.MinionConfig, proxy minion.CoordinatorProxy, args []string) {

	suite := "unstable"

	for _, archive := range args {
		needs, err := proxy.GetBuildNeeding(archive, suite, "any", "")
		if err != nil {
			log.Fatalf("%s", err)
		}
		for _, need := range needs {
			log.Printf("%s [%s] - %s", archive, need.Arch, need.Location)
			archiveRoot := fmt.Sprintf("http://%s/%s", config.Host, archive)
			dsc := fmt.Sprintf("%s/%s", archiveRoot, need.Location)
			build := minion.NewBuild(
				config.Host,
				archive,
				suite,
				"main",
				need.Arch,
				dsc,
			)
			proxy.QueueBuild(build)
		}
	}
}
