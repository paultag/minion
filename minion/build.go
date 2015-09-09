package minion

import (
	"fmt"
)

type Archive struct {
	Key      string
	Root     string
	Sections []string
	Suite    string
}

type Chroot struct {
	Target string
	Chroot string
}

type Upload struct {
	Host    string
	Port    int
	Archive string
}

type BinNMU struct {
	Changelog string
	Version   string
}

type Build struct {
	Archives []Archive
	Chroot   Chroot
	DSC      string
	Arch     string
	Upload   Upload
	BinNMU   BinNMU
}

/*
 */
func NewBuild(
	host string,
	archive string,
	suite string,
	component string,
	arch string,
	dsc string,
) Build {
	archiveRoot := fmt.Sprintf("http://%s/%s", host, archive)
	return Build{
		Archives: []Archive{
			Archive{
				Root:     archiveRoot,
				Key:      fmt.Sprintf("%s.asc", archiveRoot),
				Suite:    suite,
				Sections: []string{component},
			},
		},
		Chroot: Chroot{Chroot: suite, Target: suite},
		Upload: Upload{Host: host, Port: 1984, Archive: archive},
		Arch:   arch,
		DSC:    dsc,
	}
}

func (b Build) GetBuildChannelKey() string {
	return fmt.Sprintf("%s-%s", b.Chroot.Target, b.Arch)
}

type BuildChannelMap map[string]chan Build

func (b BuildChannelMap) Get(arch string) chan Build {
	if channel, ok := b[arch]; ok {
		return channel
	}
	b[arch] = make(chan Build, 10)
	return b[arch]
}
