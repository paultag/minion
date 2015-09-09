package minion

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
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

func ParseDscURL(url string) (*control.DSC, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf := bufio.NewReader(resp.Body)
	return control.ParseDsc(buf, "")

}

func (m *MinionRemote) Build(i Build, ftbfs *bool) error {
	chrootArch := "amd64" // XXX: FIX THIS LIKE SO HARD

	dsc, err := ParseDscURL(i.DSC)
	if err != nil {
		return err
	}

	cleanup, workdir, err := Tempdir()
	if err != nil {
		return err
	}
	// defer cleanup()
	log.Printf("%s\n", workdir, cleanup)

	/* We're in a tempdir, let's make it dirty */

	build := sbuild.NewSbuild(i.Chroot.Chroot, i.Chroot.Target)

	if i.Arch == "all" {
		build.Arch(chrootArch)
		build.AddFlag("--arch-all-only")
	} else {
		build.Arch(i.Arch)
	}
	build.BuildDepResolver("aptitude")

	buildVersion := dsc.Version

	if i.BinNMU.Version != "" {
		build.AddArgument("uploader", "Foo Bar <example@example.com>")
		build.AddArgument("maintainer", "Foo Bar <example@example.com>")
		build.AddArgument("make-binNMU", i.BinNMU.Changelog)
		build.AddArgument("binNMU", i.BinNMU.Version)

		/* In addition, let's fix buildVersion up */
		a := func(orig, v string) string {
			return fmt.Sprintf("%s+b%s", orig, v)
		}

		if buildVersion.IsNative() {
			buildVersion.Version = a(buildVersion.Version, i.BinNMU.Version)
		} else {
			buildVersion.Revision = a(buildVersion.Revision, i.BinNMU.Version)
		}
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

	// if i.Arch != chrootArch {
	//	/* Rename */
	// }

	changesFile := Filename(dsc.Source, buildVersion, i.Arch, "changes")
	logFile := Filename(dsc.Source, buildVersion, i.Arch, "build")

	err = cmd.Run()

	if i.Arch == "all" {
		/* We have to play fixup here */
		wrongChanges := Filename(dsc.Source, buildVersion, chrootArch, "changes")
		if _, err := os.Stat(wrongChanges); os.IsExist(err) {
			if err := os.Rename(
				wrongChanges,
				changesFile,
			); err != nil {
				return err
			}
		}

		if err := os.Rename(
			Filename(dsc.Source, buildVersion, chrootArch, "build"),
			logFile,
		); err != nil {
			return err
		}
	}
	if err != nil {
		changes, err := LogChangesFromDsc(logFile, *dsc, i.Chroot.Target, i.Arch)
		if err != nil {
			return err
		}
		fd, err := os.Create(changesFile)
		if err != nil {
			return err
		}
		defer fd.Close()
		_, err = fd.Write([]byte(changes))
		if err != nil {
			return err
		}
		err = UploadChanges(m.Config, i, changesFile)
		*ftbfs = true
		return err
	}

	/* Right, so we've got a complete upload, let's go ahead and dput
	 * this sucka. */
	log.Printf("Complete. Doing upload now")

	AppendLogToChanges(logFile, changesFile, i.Arch)

	log.Printf("Uploading: %s\n", changesFile)
	err = UploadChanges(m.Config, i, changesFile)
	log.Printf("Uploaded.")
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
