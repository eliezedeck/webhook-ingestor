package parameters

import (
	"flag"
)

var (
	ParamListen        = ":8080"
	ParamAdminListen   = ":8081"
	ParamAdminUsername = "admin"
	ParamAdminPassword = "admin"
)

func ParseFlags() {
	flag.StringVar(&ParamListen, "listen", ParamListen, "Address to listen as HTTP server; defaults to :8080")
	flag.StringVar(&ParamAdminListen, "admin-listen", ParamAdminListen, "Address to listen as HTTP server, for administration; defaults to :8081")
	flag.StringVar(&ParamAdminUsername, "username", ParamAdminUsername, "Username for admin; defaults to 'admin'")
	flag.StringVar(&ParamAdminPassword, "password", ParamAdminPassword, "Password for admin; defaults to 'admin'")
	flag.Parse()
}
