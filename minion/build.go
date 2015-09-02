package minion

import (
	"fmt"
)

type Archive struct {
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

type Build struct {
	Archives []Archive
	Chroot   Chroot
	DSC      string
	Arch     string
	Upload   Upload
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
