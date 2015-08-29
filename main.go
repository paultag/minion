package main

import (
	"flag"
)

var commands = []*Command{
	&coordinatorCommand,
	&remoteCommand,
	&minionCommand,
}

/* This encapsulates a Minion command to be implemented by the internals.
 * which may optionally set additional arguments. */
type Command struct {
	Name  string
	Run   func(config MinionConfig, cmd *Command, args []string)
	Flag  flag.FlagSet
	Usage string
}

func main() {
	config := GetMinionConfig()

	cert := flag.String("cert", config.Cert, "client or server .crt file")
	key := flag.String("key", config.Key, "client or server .key file")
	ca := flag.String("ca", config.CaCert, "client or server ca .crt file")
	host := flag.String("host", config.Host, "target host, or host to bind to")
	port := flag.Int("port", config.Port, "target port, or port to bind to")

	flag.Parse()

	config.Cert = *cert
	config.Key = *key
	config.CaCert = *ca
	config.Host = *host
	config.Port = *port

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}

	for _, command := range commands {
		if command.Name == args[0] {
			command.Flag.Parse(args[1:])
			args = command.Flag.Args()
			command.Run(config, command, args)
			return
		}
	}

	flag.Usage()
	return
}
