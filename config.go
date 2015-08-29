package main

import (
	"os"
	"os/user"
	"path"

	"pault.ag/go/debian/control"
)

// Model we Unmarshal the MinionRC into using the pault.ag/go/debian/control
// RFC822 Unmarshaler.
type MinionConfig struct {
	// OpenSSL Client or Server .crt to serve to the world.
	Cert string

	// OpenSSL Client or Server .key to do private key things with.
	Key string

	// OpenSSL CA Cert to compare the server or client certs to.
	CaCert string

	// Name of the Host that we're either serving on behalf of, or
	// connecting to.
	Host string

	// Port we're connecting to on the host.
	Port int

	// Mode to run as.
	Mode string
}

// Read the ~/.minionrc into a MinionConfig struct, if it exists, otherwise
// return a default empty config. This may be later overridden by CLI flags.
func GetMinionConfig() MinionConfig {
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
