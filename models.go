package main

import (
	"log"
	"net/rpc"
)

type Minion struct {
	Arches []string
}

/* */

type BuildTarget struct {
	Arch     string
	DSC      string
	Target   ChrootTarget
	Archives []Archive
}

type ChrootTarget struct {
	Chroot string
	Suite  string
}

type Archive struct {
	URL string
	Key string
}

type UpdateResult struct {
	Failed bool
}

type BuildResult struct {
	FTBFS bool
}

/* */

func (m *Minion) Build(target *BuildTarget, ret *BuildResult) error {
	log.Printf("Doing build: %s/%s\n", target.Target.Suite, target.Target.Chroot)
	log.Printf("   -> %s (%s)\n", target.DSC, target.Arch)
	return nil
}

func (m *Minion) Upgrade(target string, ret *UpdateResult) error {
	log.Printf("Doing upgrade\n")
	return nil
}

func (m *Minion) GetArches(_ interface{}, ret *[]string) error {
	*ret = m.Arches
	return nil
}

type RemoteMinion struct {
	*rpc.Client
}

func (r *RemoteMinion) Arches() ([]string, error) {
	var stub interface{}
	ret := []string{}
	err := r.Call("Minion.GetArches", &stub, &ret)
	return ret, err
}

func (r *RemoteMinion) Upgrade(target string) error {
	var stub interface{}
	return r.Call("Minion.Upgrade", &target, &stub)
}

func (r *RemoteMinion) Build(
	archives []Archive,
	target ChrootTarget,
	arch, dsc string,
) (bool, error) {
	arg := BuildTarget{
		Archives: archives,
		Target:   target,
		Arch:     arch,
		DSC:      dsc,
	}
	ret := BuildResult{FTBFS: false}
	if err := r.Call("Minion.Build", &arg, &ret); err != nil {
		return ret.FTBFS, err
	}
	return ret.FTBFS, nil
}
