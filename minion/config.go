package minion

type MinionConfig struct {
	Cert   string `flag:"cert"    description:"OpenSSL Client or Server .crt to serve to the world."`
	Key    string `flag:"key"     description:"OpenSSL Client or Server .key to do private key things with."`
	CaCert string `flag:"ca-cert" description:"OpenSSL CA Cert to compare the server or client certs to."`
	Host   string `flag:"host"    description:"Name of the Host that we're either serving on behalf of, or connecting to."`
	Port   int    `flag:"port"    description:"Port we're connecting to on the host."`
	Mode   string `flag:"mode"    description:"Mode to run as."`

	Administrator string `flag:"administrator" description:""`
	Templates     string `flag:"templates" description:"templates"`
	Repos         string `flag:"repos" description:"repos"`
}
