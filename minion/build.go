package minion

type Archive struct {
	Root     string
	Sections []string
	Suite    string
}

type Chroot struct {
	Target string
	Chroot string
}

type Build struct {
	Archives []Archive
	Chroot   Chroot
	DSC      string
	Arch     string
}

type BuildChannelMap map[string]chan Build

func (b BuildChannelMap) Get(arch string) chan Build {
	if channel, ok := b[arch]; ok {
		return channel
	}
	b[arch] = make(chan Build, 10)
	return b[arch]
}
