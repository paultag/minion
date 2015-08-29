package minion

import (
	"log"
	"net/rpc"
)

/***************************/

type MinionRemote struct {
	arches []string
}

func NewMinionRemote(arches []string) MinionRemote {
	return MinionRemote{arches: arches}
}

func (m *MinionRemote) GetArches(i *bool, ret *[]string) error {
	*ret = m.arches
	return nil
}

func (m *MinionRemote) Build(i Build, ftbfs *bool) error {
	log.Printf("Doing a build for %s\n", i)
	return nil
}

/***************************/

type MinionProxy struct {
	*rpc.Client
}

func (m *MinionProxy) GetArches() ([]string, error) {
	var c bool = true
	ret := []string{}
	return ret, m.Call("MinionRemote.GetArches", &c, &ret)
}

func (m *MinionProxy) Build(build Build) (bool, error) {
	var ftbfs bool
	return ftbfs, m.Call("MinionRemote.Build", build, &ftbfs)
}
