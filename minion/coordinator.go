package minion

import (
	"fmt"
	"log"
	"net/rpc"
	"path"

	"pault.ag/go/reprepro"
)

type BuildNeedingRequest struct {
	Repo    string
	Suite   string
	Arch    string
	Package string
}

type CoordinatorProxy struct {
	*rpc.Client
}

func (p *CoordinatorProxy) QueueBuild(build Build) error {
	var reply *error
	return p.Call("CoordinatorRemote.QueueBuild", build, &reply)
}

func (p *CoordinatorProxy) GetQueueLengths() (map[string]int, error) {
	ret := map[string]int{}
	return ret, p.Call("CoordinatorRemote.GetQueueLengths", false, &ret)
}

func (p *CoordinatorProxy) Heartbeat() ([]string, error) {
	reply := []string{}
	return reply, p.Call("CoordinatorRemote.Heartbeat", false, &reply)
}

func (p *CoordinatorProxy) GetOnlineMinions() ([]string, error) {
	reply := []string{}
	return reply, p.Call("CoordinatorRemote.GetOnlineMinions", false, &reply)
}

func (p *CoordinatorProxy) GetBuildNeeding(repo, suite, arch, pkg string) ([]reprepro.BuildNeedingPackage, error) {
	reply := []reprepro.BuildNeedingPackage{}
	return reply, p.Call("CoordinatorRemote.GetBuildNeeding", BuildNeedingRequest{
		Repo:    repo,
		Suite:   suite,
		Arch:    arch,
		Package: pkg,
	}, &reply)
}

/*
 *
 *
 *
 *
 *
 */

func NewCoordinatorRemote(
	buildChannels *BuildChannelMap,
	config *MinionConfig,
	clients *OnlineClients,
) CoordinatorRemote {
	return CoordinatorRemote{
		buildChannels: buildChannels,
		Config:        config,
		Clients:       clients,
	}
}

type CoordinatorRemote struct {
	buildChannels *BuildChannelMap
	Config        *MinionConfig
	Clients       *OnlineClients
}

func (c *CoordinatorRemote) QueueBuild(build Build, r *interface{}) error {
	log.Printf("Enqueueing build: %s\n", build)
	c.buildChannels.Get(build.GetBuildChannelKey()) <- build
	return nil
}

func (c *CoordinatorRemote) Heartbeat(incoming bool, ret *[]string) error {
	myStatus := []string{}
	for client, _ := range *c.Clients {
		err := client.Proxy.Heartbeat()
		if err != nil {
			log.Printf("%s - %s\n", client, err)
			c.Clients.Remove(client)
			continue
		}
		myStatus = append(myStatus, client.Name)
	}
	*ret = myStatus
	return nil
}

func (c *CoordinatorRemote) GetOnlineMinions(incoming bool, ret *[]string) error {
	myStatus := []string{}
	for client, _ := range *c.Clients {
		myStatus = append(myStatus, client.Name)
	}
	*ret = myStatus
	return nil
}

func (c *CoordinatorRemote) GetQueueLengths(incoming bool, ret *map[string]int) error {
	myStatus := map[string]int{}
	for key, value := range *c.buildChannels {
		myStatus[key] = len(value)
	}
	*ret = myStatus
	return nil
}

func (c *CoordinatorRemote) GetBuildNeeding(
	buildNeedingRequest BuildNeedingRequest,
	ret *[]reprepro.BuildNeedingPackage,
) error {
	repos := c.Config.Repos
	if repos == "" {
		return fmt.Errorf("Error; I don't know where repos live. Set Repos: config flag")
	}
	repo := reprepro.NewRepo(path.Join(repos, path.Clean(buildNeedingRequest.Repo)))

	var pkg *string
	if buildNeedingRequest.Package != "" {
		pkg = &buildNeedingRequest.Package
	}

	buildNeeding, err := repo.BuildNeeding(
		buildNeedingRequest.Suite,
		buildNeedingRequest.Arch,
		pkg,
	)
	if err != nil {
		return err
	}
	*ret = buildNeeding
	return nil
}
