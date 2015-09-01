package minion

import (
	"log"
	"net/rpc"
	"os"
	"os/exec"

	"pault.ag/go/sbuild"
)

/***************************/

type MinionRemote struct {
	arches []string
}

func NewMinionRemote(config MinionConfig, arches []string) MinionRemote {
	return MinionRemote{arches: arches}
}

func (m *MinionRemote) GetArches(i *bool, ret *[]string) error {
	*ret = m.arches
	return nil
}

func attachToStdout(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
}

func (m *MinionRemote) Build(i Build, ftbfs *bool) error {
	cleanup, err := Tempdir()
	if err != nil {
		return err
	}
	defer cleanup()

	/* We're in a tempdir, let's make it dirty */

	build := sbuild.NewSbuild(i.Chroot.Chroot, i.Chroot.Target)
	build.Arch(i.Arch)
	build.BuildDepResolver("aptitude")
	build.Verbose()

	cmd, err := build.BuildCommand(i.DSC)
	attachToStdout(cmd)

	if err != nil {
		return err
	}
	log.Printf("Doing a build for %s -- waiting\n", i)
	cmd.Run()
	log.Printf("Complete.")

	/* dsend this to the server target */
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
