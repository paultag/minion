package main

import (
	"flag"
	"log"
)

func main() {
	cert := flag.String("cert", "", "client or server .crt file")
	key := flag.String("key", "", "client or server .key file")
	ca := flag.String("ca", "", "client or server ca .crt file")
	mode := flag.String("mode", "", "What mode to run in (minion or coordinator)")
	host := flag.String("host", "localhost", "target host, or host to bind to")
	port := flag.Int("port", 8765, "target port, or port to bind to")

	flag.Parse()

	switch *mode {
	case "minion":
		BeAMinion(*cert, *key, *ca, *host, *port)
	case "coordinator":
		BeACoordinator(*cert, *key, *ca, *host, *port)
	default:
		log.Fatalf("Don't know what to do :(\n")
	}
}
