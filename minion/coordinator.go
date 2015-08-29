package minion

import (
	"log"
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

func NewCoordinatorRemote(buildChannels *map[string]chan Build) CoordinatorRemote {
	return CoordinatorRemote{buildChannels: buildChannels}
}

type CoordinatorRemote struct {
	buildChannels *map[string]chan Build
}

func (c *CoordinatorRemote) QueueBuild(build Build, r *interface{}) error {
	log.Printf("Queueing %s\n", build.Arch)
	(*c.buildChannels)[build.Arch] <- build
	log.Printf("Queued\n")
	return nil
}
