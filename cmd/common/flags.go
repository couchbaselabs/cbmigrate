package common

import "github.com/couchbaselabs/cbmigrate/cmd/flag"

const (
	CBCluster            = "cb-cluster"
	CBUsername           = "cb-username"
	CBPassword           = "cb-password"
	CBClientCert         = "cb-client-cert"
	CBClientCertPassword = "cb-client-cert-password"
	CBClientKey          = "cb-client-key"
	CBClientKeyPassword  = "cb-client-key-password"
	CBGenerateKey        = "cb-generate-key"
	CBCACert             = "cb-cacert"
	CBNoSSLVerify        = "cb-no-ssl-verify"
	CBBucket             = "cb-bucket"
	CBScope              = "cb-scope"
	CBCollection         = "cb-collection"
	CBBatchSize          = "cb-batch-size"

	CopyIndexes     = "copy-indexes"
	BufferSize      = "buffer-size"
	HashDocumentKey = "hash-document-key"
)

var cbCluster = &flag.StringFlag{
	Name:     CBCluster,
	Usage:    "The hostname of a node in the cluster to import data into.",
	Required: true,
}

var cbUsername = &flag.StringFlag{
	Name:     CBUsername,
	Usage:    "The username for cluster authentication.",
	Required: true,
}

var cbPassword = &flag.StringFlag{
	Name:     CBPassword,
	Usage:    "The password for cluster authentication.",
	Required: true,
}

var cbClientCert = &flag.StringFlag{
	Name: CBClientCert,
	Usage: "The path to a client certificate used to authenticate when connecting to a cluster. " +
		"May be supplied with --client-key as an alternative to the --username and --password flags.",
	Required: true,
}

var cbClientCertPassword = &flag.StringFlag{
	Name:  CBClientCertPassword,
	Usage: "The password for the certificate provided to the --client-cert flag, when using this flag, the certificate/key pair is expected to be in the PKCS#12 format.",
}

var cbClientKey = &flag.StringFlag{
	Name: CBClientKey,
	Usage: "The path to the client private key whose public key is contained in the certificate provided to the --client-cert flag." +
		" May be supplied with --client-cert as an alternative to the --username and --password flags.",
}

var cbClientKeyPassword = &flag.StringFlag{
	Name:  CBClientKeyPassword,
	Usage: "The password for the key provided to the --client-key flag, when using this flag, the key is expected to be in the PKCS#8 format.",
}

var hashDocumentKey = &flag.EnumFlag{
	Name:   HashDocumentKey,
	Usage:  "Hash the couchbase document key.",
	Values: []string{"sha256", "sha512"},
}

func GetCBGenerateKeyOption(value string) flag.Flag {
	return &flag.StringFlag{
		Name: CBGenerateKey,
		Usage: "Specifies a key expression used for generating a key for each document imported." +
			" This option allows for the creation of unique document keys in Couchbase by combining static text," +
			" field values (denoted by %fieldname%), and custom generators (like #UUID#) in a format like \"key::%name%::#UUID#\"",
		Value: value,
	}
}

var cbCACert = &flag.StringFlag{
	Name: CBCACert,
	Usage: "Specifies a CA certificate that will be used to verify the identity of the server being connecting to. " +
		"Either this flag or the --no-ssl-verify flag must be specified when using an SSL encrypted connection.",
}

var cbNoSSLVerify = &flag.StringFlag{
	Name: CBNoSSLVerify,
	Usage: "Skips the SSL verification phase. Specifying this flag will allow a connection using SSL encryption, " +
		"but will not verify the identity of the server you connect to. " +
		"You are vulnerable to a man-in-the-middle attack if you use this flag." +
		" Either this flag or the --cacert flag must be specified when using an SSL encrypted connection.",
}

var cbBucket = &flag.StringFlag{
	Name:  CBBucket,
	Usage: "The name of the couchbase bucket.",
}

var cbScope = &flag.StringFlag{
	Name:  CBScope,
	Usage: "The name of the scope in which the collection resides. If the scope does not exist, it will be created.",
}

var cbCollection = &flag.StringFlag{
	Name:  CBCollection,
	Usage: "The name of the collection where the data needs to be imported. If the collection does not exist, it will be created.",
}

var batchSize = &flag.IntFlag{
	Name:  CBBatchSize,
	Usage: "Batch size",
	Value: 200,
}

var copyIndexes = &flag.BoolFlag{
	Name:  CopyIndexes,
	Usage: "Copy indexes for the collection",
	Value: true,
}

var bufferSize = &flag.IntFlag{
	Name:  BufferSize,
	Usage: "Buffer size",
	Value: 10000,
}

func GetCBFlags() []flag.Flag {
	flags := []flag.Flag{
		cbCluster,
		&flag.CompositeFlag{
			Flags: []flag.Flag{
				&flag.CompositeFlag{
					Flags: []flag.Flag{
						cbUsername,
						cbPassword,
					},
					Required: true,
				},
				&flag.CompositeFlag{
					Flags: []flag.Flag{
						cbClientCert,
						cbClientCertPassword,
						cbClientKey,
						cbClientKeyPassword,
					},
					Required: true,
				},
			},
			Type:          flag.RelationshipOR,
			RequiredBrace: true,
			Required:      true,
		},
		cbCACert,
		cbNoSSLVerify,
		cbBucket,
		cbScope,
		cbCollection,
		batchSize,
		hashDocumentKey,
		GetDebugFlag(),
	}
	return flags
}

func GetCommonFlags() []flag.Flag {
	return []flag.Flag{
		copyIndexes,
		bufferSize,
	}
}
