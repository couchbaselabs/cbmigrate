package command

import (
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/flag"
	"github.com/couchbaselabs/cbmigrate/internal/feature"

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
	Hidden:   !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBPort = &flag.StringFlag{
	Name:     MongoDBPort,
	Usage:    "server port (can also use --host hostname:port)",
	Required: true,
	Hidden:   !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBSSL = &flag.StringFlag{
	Name:   MongoDBSSL,
	Usage:  "connect to a mongod or mongos that has ssl enabled",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBSSLCAFile = &flag.StringFlag{
	Name:   MongoDBSSLCAFile,
	Usage:  "the .pem file containing the root certificate chain from the certificate authority",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBSSLPEMKeyFile = &flag.StringFlag{
	Name:   MongoDBSSLPEMKeyFile,
	Usage:  "the .pem file containing the certificate and key",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBSSLPEMKeyPassword = &flag.StringFlag{
	Name:   MongoDBSSLPEMKeyPassword,
	Usage:  "the password to decrypt the sslPEMKeyFile, if necessary",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBSSLCRLFile = &flag.StringFlag{
	Name:   MongoDBSSLCRLFile,
	Usage:  "the .pem file containing the certificate revocation list",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBSSLFIPSMode = &flag.StringFlag{
	Name:   MongoDBSSLFIPSMode,
	Usage:  "use FIPS mode of the installed openssl library",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBTLSInsecure = &flag.StringFlag{
	Name:   MongoDBTLSInsecure,
	Usage:  "bypass the validation for server's certificate chain and host name",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBUsername = &flag.StringFlag{
	Name:   MongoDBUsername,
	Usage:  "username for authentication",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBPassword = &flag.StringFlag{
	Name:   MongoDBPassword,
	Usage:  "password for authentication",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBAuthDatabase = &flag.StringFlag{
	Name:   MongoDBAuthDatabase,
	Usage:  "database that holds the user's credentials",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBAuthMechanism = &flag.StringFlag{
	Name:   MongoDBAuthMechanism,
	Usage:  "authentication mechanism to use",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBAWSSessionToken = &flag.StringFlag{
	Name:   MongoDBAWSSessionToken,
	Usage:  "session token to authenticate via AWS IAM",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBGSSAPIServiceName = &flag.StringFlag{
	Name:   MongoDBGSSAPIServiceName,
	Usage:  "service name to use when authenticating using GSSAPI/Kerberos (default: mongodb)",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBGSSAPIHostName = &flag.StringFlag{
	Name:   MongoDBGSSAPIHostName,
	Usage:  "hostname to use when authenticating using GSSAPI/Kerberos (default: <remote server's address>)",
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
}

var mongoDBDatabase = &flag.StringFlag{
	Name:     MongoDBDatabase,
	Usage:    "database to use",
	Required: true,
}

var mongoDBCollection = &flag.StringFlag{
	Name:     MongoDBCollection,
	Usage:    "collection to use",
	Required: true,
}

var mongoDBURI = &flag.StringFlag{
	Name:     MongoDBURI,
	Usage:    "mongodb uri connection string",
	Required: true,
}

var mongoDBReadPreference = &flag.StringFlag{
	Name:   MongoDBReadPreference,
	Usage:  `specify either a preference mode (e.g. 'nearest') or a preference json object (e.g. '{mode: "nearest", tagSets: [{a: "b"}],  maxStalenessSeconds: 123}')`,
	Hidden: !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
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
					},
					Hidden:   !feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
					Required: true,
				},
			},
			Required:      true,
			RequiredBrace: feature.IsFeatureEnabled(feature.CbmigrateMongoHostOptsConfig),
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
		mongoDBDatabase,
		mongoDBReadPreference,
	}
	flags = append(flags, common.GetCBFlags()...)
	flags = append(flags, common.GetCommonFlags()...)
	examples := []common.Example{
		{
			Value: "cbmigrate mongo --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name",
			Usage: "Imports data from MongoDB to Couchbase, using the MongoDB collection name as the Couchbase collection name. The default generator key is set to %_id%, leveraging MongoDB's unique identifier for each document.",
		},
		{
			Value: "cbmigrate mongo --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::%name.first_name%::%name.last_name%",
			Usage: "Imports the data from mongo to couchbase, allowing the use of dot notation (e.g., name.first_name) to reference nested fields. However, it does not support referencing fields within an array of documents.",
		},
		{
			Value: "cbmigrate mongo --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::#UUID#",
			Usage: "Imports the data from mongo to couchbase.",
		},
	}
	return common.NewCommand("mongo", []string{"m"}, examples, "", "", flags)
}
