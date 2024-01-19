package command

import (
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/flag"

	"github.com/spf13/cobra"
)

const (
	MongoDBHost              = "mongodb-host"
	MongoDBPort              = "mongodb-port"
	MongoDBSSL               = "mongodb-ssl"
	MongoDBSSLCAFile         = "mongodb-ssl-ca-file"
	MongoDBSSLPEMKeyFile     = "mongodb-ssl-pem-key-file"
	MongoDBSSLPEMKeyPassword = "mongodb-ssl-pem-key-password"
	MongoDBSSLCRLFile        = "mongodb-ssl-crl-file"
	MongoDBSSLFIPSMode       = "mongodb-ssl-fips-mode"
	MongoDBTLSInsecure       = "mongodb-tls-insecure"
	MongoDBUsername          = "mongodb-username"
	MongoDBPassword          = "mongodb-password"
	MongoDBAuthDatabase      = "mongodb-authentication-database"
	MongoDBAuthMechanism     = "mongodb-authentication-mechanism"
	MongoDBAWSSessionToken   = "mongodb-aws-session-token"
	MongoDBGSSAPIServiceName = "mongodb-gss-api-service-name"
	MongoDBGSSAPIHostName    = "mongodb-gss-api-host-name"
	MongoDBDatabase          = "mongodb-database"
	MongoDBCollection        = "mongodb-collection"
	MongoDBURI               = "mongodb-uri"
	MongoDBReadPreference    = "mongodb-read-preference"
)

var mongoDBHost = &flag.StringFlag{
	Name:     MongoDBHost,
	Usage:    "mongodb host to connect to (setname/host1,host2 for replica sets)",
	Required: true,
}

var mongoDBPort = &flag.StringFlag{
	Name:     MongoDBPort,
	Usage:    "server port (can also use --host hostname:port)",
	Required: true,
}

var mongoDBSSL = &flag.StringFlag{
	Name:  MongoDBSSL,
	Usage: "connect to a mongod or mongos that has ssl enabled",
}

var mongoDBSSLCAFile = &flag.StringFlag{
	Name:  MongoDBSSLCAFile,
	Usage: "the .pem file containing the root certificate chain from the certificate authority",
}

var mongoDBSSLPEMKeyFile = &flag.StringFlag{
	Name:  MongoDBSSLPEMKeyFile,
	Usage: "the .pem file containing the certificate and key",
}

var mongoDBSSLPEMKeyPassword = &flag.StringFlag{
	Name:  MongoDBSSLPEMKeyPassword,
	Usage: "the password to decrypt the sslPEMKeyFile, if necessary",
}

var mongoDBSSLCRLFile = &flag.StringFlag{
	Name:  MongoDBSSLCRLFile,
	Usage: "the .pem file containing the certificate revocation list",
}

var mongoDBSSLFIPSMode = &flag.StringFlag{
	Name:  MongoDBSSLFIPSMode,
	Usage: "use FIPS mode of the installed openssl library",
}

var mongoDBTLSInsecure = &flag.StringFlag{
	Name:  MongoDBTLSInsecure,
	Usage: "bypass the validation for server's certificate chain and host name",
}

var mongoDBUsername = &flag.StringFlag{
	Name:  MongoDBUsername,
	Usage: "username for authentication",
}

var mongoDBPassword = &flag.StringFlag{
	Name:  MongoDBPassword,
	Usage: "password for authentication",
}

var mongoDBAuthDatabase = &flag.StringFlag{
	Name:  MongoDBAuthDatabase,
	Usage: "database that holds the user's credentials",
}

var mongoDBAuthMechanism = &flag.StringFlag{
	Name:  MongoDBAuthMechanism,
	Usage: "authentication mechanism to use",
}

var mongoDBAWSSessionToken = &flag.StringFlag{
	Name:  MongoDBAWSSessionToken,
	Usage: "session token to authenticate via AWS IAM",
}

var mongoDBGSSAPIServiceName = &flag.StringFlag{
	Name:  MongoDBGSSAPIServiceName,
	Usage: "service name to use when authenticating using GSSAPI/Kerberos (default: mongodb)",
}

var mongoDBGSSAPIHostName = &flag.StringFlag{
	Name:  MongoDBGSSAPIHostName,
	Usage: "hostname to use when authenticating using GSSAPI/Kerberos (default: <remote server's address>)",
}

var mongoDBDatabase = &flag.StringFlag{
	Name:  MongoDBDatabase,
	Usage: "database to use",
}

var mongoDBCollection = &flag.StringFlag{
	Name:  MongoDBCollection,
	Usage: "collection to use",
}

var mongoDBURI = &flag.StringFlag{
	Name:     MongoDBURI,
	Usage:    "mongodb uri connection string",
	Required: true,
}

var mongoDBReadPreference = &flag.StringFlag{
	Name:  MongoDBReadPreference,
	Usage: `specify either a preference mode (e.g. 'nearest') or a preference json object (e.g. '{mode: "nearest", tagSets: [{a: "b"}],  maxStalenessSeconds: 123}')`,
}

func NewCommand() *cobra.Command {

	//short := "A tool to convert time series data in CSV to the one supported by Couchbase."
	//long := `cbmigrate mongo is a CLI tool for Couchbase that enables users to convert time series data in CSV to the format required by Couchbase.`
	flags := []flag.Flag{
		&flag.CompositeFlag{
			Flags: []flag.Flag{
				mongoDBURI,
				&flag.CompositeFlag{
					Flags: []flag.Flag{
						mongoDBHost,
						mongoDBPort,
						mongoDBSSL,
						mongoDBUsername,
						mongoDBPassword,
						mongoDBAuthDatabase,
						mongoDBAuthMechanism,
						mongoDBDatabase,
					},
				},
			},
			RequiredBrace: true,
			Type:          flag.RelationshipOR,
		},
		mongoDBSSLCAFile,
		mongoDBSSLPEMKeyFile,
		mongoDBSSLPEMKeyPassword,
		mongoDBSSLCRLFile,
		mongoDBSSLFIPSMode,
		mongoDBTLSInsecure,
		mongoDBAWSSessionToken,
		mongoDBGSSAPIServiceName,
		mongoDBGSSAPIHostName,
		mongoDBCollection,
		mongoDBReadPreference,
	}
	flags = append(flags, common.GetCBFlags()...)
	examples := []common.Example{
		{
			Value: "cbmigrate mongo ",
		},
	}
	return common.NewCommand("mongo", []string{"m"}, examples, "", "", flags)
}
