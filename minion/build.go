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
