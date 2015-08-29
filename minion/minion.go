package minion

import (
	"net/rpc"
)

type MinionRemote struct {
	Arches []string
}

func (m *MinionRemote) GetArches(i *bool, ret *[]string) error {
	*ret = m.Arches
	return nil
}

type MinionProxy struct {
	*rpc.Client
}

func (m *MinionProxy) GetArches() ([]string, error) {
	var c bool = true
	ret := []string{}
	return ret, m.Call("MinionRemote.GetArches", &c, &ret)
}
