package config

var (
	DataDir = "/var/lib/gocockpit"

	WebDir  = DataDir + "/web"
	CertDir = DataDir + "/cert"

	ConfigFile = "/etc/gocockpit/gocockpit.toml"
)