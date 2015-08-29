package minion

type Archive struct {
	Root     string
	Sections []string
	Suite    string
}

type Build struct {
	Archives []Archive
	DSC      string
	Arch     string
}

type BuildChannelMap map[string]chan Build

func (b BuildChannelMap) Get(arch string) chan Build {
	if channel, ok := b[arch]; ok {
		return channel
	}
	b[arch] = make(chan Build)
	return b[arch]
}
