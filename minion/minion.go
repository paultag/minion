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
		build.BuildArch("all")
		build.AddFlag("--arch-all-only")
	} else {
		build.Arch(i.Arch)
	}
	build.BuildDepResolver("aptitude")

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
	cmd.Run()
	/* set ftbfs here */
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
