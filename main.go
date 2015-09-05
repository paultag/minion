package main

import (
	"flag"
	"os"

	"pault.ag/go/config"
	"pault.ag/go/minion/minion"
)

var commands = []*Command{
	&coordinatorCommand,
	&repreproCommand,
	&minionCommand,
}

/* This encapsulates a Minion command to be implemented by the internals.
 * which may optionally set additional arguments. */
type Command struct {
	Name  string
	Run   func(conf minion.MinionConfig, cmd *Command, args []string)
	Flag  flag.FlagSet
	Usage string
}

func main() {
	conf := minion.MinionConfig{
		Host: "localhost",
		Mode: "minion",
		Port: 8765,
	}

	flags, err := config.LoadFlags("minion", &conf)
	if err != nil {
		panic(err)
	}

	flags.Parse(os.Args[1:])

	args := flags.Args()
	if len(args) == 0 {
		flags.Usage()
		return
	}

	for _, command := range commands {
		if command.Name == args[0] {
			command.Flag.Parse(args[1:])
			args = command.Flag.Args()
			command.Run(conf, command, args)
			return
		}
	}

	flags.Usage()
	return
}
