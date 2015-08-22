package main

type Minion struct {
	Arches []string
}

func (m *Minion) GetArches(_ interface{}, ret *[]string) error {
	*ret = m.Arches
	return nil
}
