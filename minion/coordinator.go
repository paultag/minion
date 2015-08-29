package minion

import (
	"net/rpc"
)

type CoordinatorProxy struct {
	*rpc.Client
}

func (p *CoordinatorProxy) QueueBuild(build Build) error {
	var reply *error
	return p.Call("CoordinatorRemote.QueueBuild", build, &reply)
}

/*
 *
 *
 *
 *
 *
 */

func NewCoordinatorRemote(buildChannels *BuildChannelMap) CoordinatorRemote {
	return CoordinatorRemote{buildChannels: buildChannels}
}

type CoordinatorRemote struct {
	buildChannels *BuildChannelMap
}

func (c *CoordinatorRemote) QueueBuild(build Build, r *interface{}) error {
	c.buildChannels.Get(build.Arch) <- build
	return nil
}
