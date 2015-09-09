package minion

type OnlineClient struct {
	Name  string
	Proxy *MinionProxy
}

type OnlineClients map[*OnlineClient]bool

func (o OnlineClients) Remove(client *OnlineClient) {
	delete(o, client)
}

func (o OnlineClients) Add(client *OnlineClient) {
	o[client] = true
}
