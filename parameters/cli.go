package parameters

import (
	"flag"
	"os"

	"github.com/eliezedeck/gobase/logging"
	"go.uber.org/zap"
)

var (
	ParamListen = ":8080"

	ParamAdminListen   = ":8081"
	ParamAdminUsername = "admin"
	ParamAdminPassword = "admin"
	ParamAdminPath     = "__admin__"

	ParamStorage         = "memory"
	ParamStorageMongoUri = "mongodb://localhost:27017"
	ParamStorageMongoDb  = "webhook-ingestor"
)

func ParseFlags() {
	flag.StringVar(&ParamListen, "listen", ParamListen, "Address to listen as HTTP server; defaults to :8080")
	flag.StringVar(&ParamAdminListen, "admin-listen", ParamAdminListen, "Address to listen as HTTP server, for administration; defaults to :8081")
	flag.StringVar(&ParamAdminUsername, "username", ParamAdminUsername, "Username for admin; defaults to 'admin'")
	flag.StringVar(&ParamAdminPassword, "password", ParamAdminPassword, "Password for admin; defaults to 'admin'")
	flag.StringVar(&ParamAdminPath, "admin-path", ParamAdminPath, "Path for admin; defaults to '__admin__'")
	flag.StringVar(&ParamStorage, "storage", ParamStorage, "Storage type; defaults to 'memory'")
	flag.StringVar(&ParamStorageMongoUri, "mongo-uri", ParamStorageMongoUri, "MongoDB URI; defaults to 'mongodb://localhost:27017'")
	flag.StringVar(&ParamStorageMongoDb, "mongo-db", ParamStorageMongoDb, "MongoDB database to use; defaults to 'webhook-ingestor'")
	flag.Parse()

	if ParamStorageMongoUri == "MONGO_URI" {
		ParamStorageMongoUri = os.Getenv("MONGO_URI")
		logging.L.Info("Using MONGO_URI from the environment", zap.String("uri", ParamStorageMongoUri))
	}
}
