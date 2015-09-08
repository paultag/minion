package minion

import (
	"fmt"
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
) CoordinatorRemote {
	return CoordinatorRemote{buildChannels: buildChannels, Config: config}
}

type CoordinatorRemote struct {
	buildChannels *BuildChannelMap
	Config        *MinionConfig
}

func (c *CoordinatorRemote) QueueBuild(build Build, r *interface{}) error {
	c.buildChannels.Get(build.GetBuildChannelKey()) <- build
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
