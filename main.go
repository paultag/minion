package main

import (
	"flag"
	"log"
	"os"
	"os/user"
	"path"

	"pault.ag/go/debian/control"
)

func getMinionRC() MinionConfig {
	ret := MinionConfig{
		Host: "localhost",
		Mode: "minion",
		Port: 8765,
	}
	localUser, err := user.Current()
	if err != nil {
		return ret
	}
	rcPath := path.Join(localUser.HomeDir, ".minionrc")
	fd, err := os.Open(rcPath)
	if err != nil {
		return ret
	}
	err = control.Unmarshal(&ret, fd)
	if err != nil {
		/* Here so I remember to add more robust debug */
		return ret
	}
	return ret
}

func main() {
	minionRC := getMinionRC()

	cert := flag.String("cert", minionRC.Cert, "client or server .crt file")
	key := flag.String("key", minionRC.Key, "client or server .key file")
	ca := flag.String("ca", minionRC.CaCert, "client or server ca .crt file")
	mode := flag.String("mode", minionRC.Mode, "What mode to run in (minion or coordinator)")
	host := flag.String("host", minionRC.Host, "target host, or host to bind to")
	port := flag.Int("port", minionRC.Port, "target port, or port to bind to")

	flag.Parse()

	minionRC.Cert = *cert
	minionRC.Key = *key
	minionRC.CaCert = *ca
	minionRC.Mode = *mode
	minionRC.Host = *host
	minionRC.Port = *port

	switch *mode {
	case "minion":
		BeAMinion(minionRC)
	case "coordinator":
		BeACoordinator(minionRC)
	case "remote":
		BeARemote(minionRC)
	default:
		log.Fatalf("Don't know what to do :(\n")
	}
}

// vim: foldmethod=marker
