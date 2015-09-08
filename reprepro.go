package main

import (
	"fmt"
	"log"
	"path"
	"strings"

	"pault.ag/go/minion/minion"
	"pault.ag/go/reprepro"
	"pault.ag/go/service"
)

var archive *string
var fqdn *string

var repreproCommand = Command{
	Name:  "reprepro",
	Run:   repreproRun,
	Usage: ``,
}

func init() {
	archive = repreproCommand.Flag.String("archive", "", "archive we're hacking on")
	fqdn = repreproCommand.Flag.String("fqdn", "", "root fqdn")
}

type Incoming struct {
	Type      string
	Suite     string
	Flavor    string
	Component string
	Arch      string
	Package   string
	Version   string
	Files     []string
}

func (i *Incoming) Parse(args []string) error {
	if len(args) < 8 {
		return fmt.Errorf("Malformed request: %s", args)
	}
	i.Type = args[0]
	i.Suite = args[1]
	i.Flavor = args[2]
	i.Component = args[3]
	i.Arch = args[4]
	i.Package = args[5]
	i.Version = args[6]

	for _, pkg := range args[8:] {
		i.Files = append(i.Files, pkg)
	}

	return nil
}

func (i *Incoming) GetDSC() (string, error) {
	if i.Flavor != "dsc" {
		return "", fmt.Errorf("Flavor is '%s', not dsc", i.Flavor)
	}
	for _, pkg := range i.Files {
		if strings.HasSuffix(pkg, ".dsc") {
			return pkg, nil
		}
	}
	return "", fmt.Errorf("No such file D:")
}

func repreproRun(config minion.MinionConfig, cmd *Command, args []string) {
	incoming := Incoming{}
	incoming.Parse(args)
	if incoming.Type != "add" {
		return
	}

	dscPath, err := incoming.GetDSC()
	if err != nil {
		log.Fatalf("Error! %s\n", err)
	}
	log.Printf("%s\n", dscPath)

	conn, err := service.DialFromKeys(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		config.Cert, config.Key, config.CaCert,
	)
	if err != nil {
		log.Fatalf("Error! %s\n", err)
	}
	proxy := minion.CoordinatorProxy{service.Client(conn)}
	log.Printf("Queueing build\n")

	archiveRoot := fmt.Sprintf("http://%s", path.Join(*fqdn, *archive))

	repo := reprepro.GetWorkingRepo()
	buildNeeding, err := repo.BuildNeeding(incoming.Suite, "any", &incoming.Package)
	if err != nil {
		log.Fatalf("Error! %s\n", err)
	}

	for _, build := range buildNeeding {
		QueueBuildNeeding(proxy, archiveRoot, build, incoming.Suite,
			incoming.Component, dscPath, *fqdn, *archive)
	}
	log.Printf("Queued\n")
}

func QueueBuildNeeding(
	proxy minion.CoordinatorProxy,
	archiveRoot string,
	build reprepro.BuildNeedingPackage,
	suite string,
	component string,
	dscPath string,
	uploadHost string,
	uploadArchive string,
) {
	proxy.QueueBuild(minion.Build{
		Archives: []minion.Archive{
			minion.Archive{
				Root:     archiveRoot,
				Key:      fmt.Sprintf("%s.asc", archiveRoot),
				Suite:    suite,
				Sections: []string{component},
			},
		},
		Chroot: minion.Chroot{
			Chroot: suite,
			Target: suite,
		},
		Arch: build.Arch,
		DSC:  fmt.Sprintf("%s/%s", archiveRoot, dscPath),
		Upload: minion.Upload{
			Host:    uploadHost,
			Port:    1984,
			Archive: uploadArchive,
		},
	})
}
