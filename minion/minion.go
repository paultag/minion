package minion

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"pault.ag/go/debian/control"
	"pault.ag/go/descend/descend"
	"pault.ag/go/sbuild"
)

/***************************/

type MinionRemote struct {
	BuildableSuites []BuildableSuite
	Config          MinionConfig
}

type BuildableSuite struct {
	Suite string
	Arch  string
}

func (b *BuildableSuite) GetKey() string {
	return fmt.Sprintf("%s-%s", b.Suite, b.Arch)
}

func NewMinionRemote(config MinionConfig, suites []BuildableSuite) MinionRemote {
	return MinionRemote{BuildableSuites: suites, Config: config}
}

func (m *MinionRemote) GetBuildableSuites(i *bool, ret *[]BuildableSuite) error {
	*ret = m.BuildableSuites
	return nil
}

func (m *MinionRemote) Heartbeat(i *bool, ret *bool) error {
	return nil
}

func attachToStdout(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
}

func (m *MinionRemote) Build(i Build, ftbfs *bool) error {
	cleanup, workdir, err := Tempdir()
	if err != nil {
		return err
	}
	defer cleanup()

	/* We're in a tempdir, let's make it dirty */

	build := sbuild.NewSbuild(i.Chroot.Chroot, i.Chroot.Target)
	if i.Arch == "all" {
		build.AddFlag("--arch-all-only")
	} else {
		build.Arch(i.Arch)
	}
	build.BuildDepResolver("aptitude")

	if i.BinNMU.Version != "" {
		build.AddArgument("uploader", "Foo Bar <example@example.com>")
		build.AddArgument("maintainer", "Foo Bar <example@example.com>")
		build.AddArgument("make-binNMU", i.BinNMU.Changelog)
		build.AddArgument("binNMU", i.BinNMU.Version)
	}

	for _, archive := range i.Archives {
		if archive.Key != "" {
			cleanup, archiveKey, err := Download(archive.Key)
			if err != nil {
				return err
			}
			defer cleanup()
			build.AddArgument("extra-repository-key", archiveKey)
		}

		build.AddArgument("extra-repository", fmt.Sprintf(
			"deb %s %s %s",
			archive.Root,
			archive.Suite,
			strings.Join(archive.Sections, " "),
		))
	}

	build.Verbose()
	cmd, err := build.BuildCommand(i.DSC)
	attachToStdout(cmd)

	if err != nil {
		return err
	}
	log.Printf("Doing a build for %s -- waiting\n", i)
	err = cmd.Run()
	if err != nil {
		*ftbfs = true
		return nil
	}

	/* Right, so we've got a complete upload, let's go ahead and dput
	 * this sucka. */
	log.Printf("Complete. Doing upload now")

	/* dsend this to the server target */
	files, err := filepath.Glob(path.Join(workdir, "*changes"))
	if err != nil {
		return err
	}

	for _, changesFile := range files {
		log.Printf("Uploading: %s\n", changesFile)
		err = UploadChanges(m.Config, i, changesFile)
		log.Printf("Uploaded.")
	}
	return err
}

func UploadChanges(conf MinionConfig, job Build, changesPath string) error {
	client, err := descend.NewClient(conf.CaCert, conf.Cert, conf.Key)
	if err != nil {
		return err
	}

	changes, err := control.ParseChangesFile(changesPath)
	if err != nil {
		return err
	}

	err = descend.DoPutChanges(
		client, changes,
		fmt.Sprintf("%s:%d", job.Upload.Host, job.Upload.Port),
		job.Upload.Archive,
	)
	return err
}

/***************************/

type MinionProxy struct {
	*rpc.Client
}

func (m *MinionProxy) GetBuildableSuites() ([]BuildableSuite, error) {
	var c bool = true
	ret := []BuildableSuite{}
	return ret, m.Call("MinionRemote.GetBuildableSuites", &c, &ret)
}

func (m *MinionProxy) Build(build Build) (bool, error) {
	var ftbfs bool
	return ftbfs, m.Call("MinionRemote.Build", build, &ftbfs)
}

func (m *MinionProxy) Heartbeat() error {
	var alive bool
	return m.Call("MinionRemote.Heartbeat", false, &alive)
}
