package minion

import (
	"log"
	"net/rpc"
	"os"

	"pault.ag/go/sbuild"
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
	cleanup, err := Tempdir()
	if err != nil {
		return err
	}
	defer cleanup()

	build := sbuild.NewSbuild(i.Chroot.Chroot, i.Chroot.Target)
	build.Arch(i.Arch)
	build.BuildDepResolver("aptitude")
	build.Verbose()

	cmd, err := build.BuildCommand(i.DSC)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err != nil {
		return err
	}
	log.Printf("Doing a build for %s -- waiting\n", i)
	cmd.Run()
	log.Printf("Complete.")
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
