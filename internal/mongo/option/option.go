package option

import (
	"encoding/pem"
	"fmt"
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/mongo/command"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"os"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	*URI
	*General
	*Connection
	*SSL
	*Auth
	*Kerberos
	*Namespace
	// Force direct connection to the server and disable the
	// drivers automatic repl set discovery logic.
	Direct bool

	// ReplicaSetName, if specified, will prevent the obtained session from
	// communicating with any server which is not part of a replica set
	// with the given name. The default is to communicate with any server
	// specified or discovered via the servers contacted.
	ReplicaSetName string

	// ReadPreference, if specified, sets the client default
	ReadPreference *readpref.ReadPref

	// WriteConcern, if specified, sets the client default
	WriteConcern *writeconcern.WriteConcern

	// RetryWrites, if specified, sets the client default.
	RetryWrites *bool

	// Query options
	QueryOptions

	CopyIndexes bool
}

// NormalizeOptionsAndURI syncs the connection string and toolOptions objects.
// It returns an error if there is any conflict between options and the connection string.
// If a value is set on the options, but not the connection string, that value is added to the
// connection string. If a value is set on the connection string, but not the options,
// that value is added to the options.
func (opts *Options) NormalizeOptionsAndURI() error {
	if opts.URI == nil || opts.URI.ConnectionString == "" {
		// If URI not provided, get replica set name and generate connection string
		_, opts.ReplicaSetName = common.SplitHostArg(opts.Host)
		uri, err := NewURI(common.BuildURI(opts.Host, opts.Port))
		if err != nil {
			return err
		}
		opts.URI = uri
	}

	cs, err := connstring.Parse(opts.URI.ConnectionString)
	if err != nil {
		return err
	}
	err = opts.setOptionsFromURI(*cs)
	if err != nil {
		return err
	}

	// finalize auth options, filling in missing passwords
	if opts.Auth.ShouldAskForPassword() {
		return fmt.Errorf("password missing for the user")
	}

	shouldAskForSSLPassword, err := opts.SSL.ShouldAskForPassword()
	if err != nil {
		return fmt.Errorf("error determining whether client cert needs password: %v", err)
	}
	if shouldAskForSSLPassword {
		return fmt.Errorf("password missing for the client cert")
	}

	err = opts.ConnString.Validate()
	if err != nil {
		return errors.Wrap(err, "connection string failed validation")
	}

	// Connect directly to a host if there's no replica set specified, or
	// if the connection string already specified a direct connection.
	// Do not connect directly if loadbalanced.
	if !opts.ConnString.LoadBalanced {
		opts.Direct = (opts.ReplicaSetName == "") || opts.Direct
	}

	return nil
}

// Sets options from the URI. If any options are already set, they are added to the connection string.
// which is eventually added to the connString field.
// Most CLI and URI options are normalized in three steps:
//
// 1. If both CLI option and URI option are set, throw an error if they conflict.
// 2. If the CLI option is set, but the URI option isn't, set the URI option
// 3. If the URI option is set, but the CLI option isn't, set the CLI option
//
// Some options (e.g. host and port) are more complicated. To check if a CLI option is set,
// we check that it is not equal to its default value. To check that a URI option is set,
// some options have an "OptionSet" field.
func (opts *Options) setOptionsFromURI(cs connstring.ConnString) error {
	opts.URI.ConnString = cs

	if opts.Port != "" {
		// if --port is set, check that each host:port pair in the URI the port defined in --port
		for i, host := range cs.Hosts {
			if strings.Index(host, ":") != -1 {
				hostPort := strings.Split(host, ":")[1]
				if hostPort != opts.Port {
					return ConflictingArgsErrorFormat("port", strings.Join(cs.Hosts, ","), opts.Port, command.MongoDBPort)
				}
			} else {
				// if the URI hosts have no ports, append them
				cs.Hosts[i] = cs.Hosts[i] + ":" + opts.Port
			}
		}
	}

	if opts.Host != "" {
		// build hosts from --host and --port
		seedlist, replicaSetName := common.SplitHostArg(opts.Host)
		opts.ReplicaSetName = replicaSetName

		if opts.Port != "" {
			for i := range seedlist {
				if strings.Index(seedlist[i], ":") == -1 { // no port
					seedlist[i] = seedlist[i] + ":" + opts.Port
				}
			}
		}

		// create a set of hosts since the order of a seedlist doesn't matter
		csHostSet := make(map[string]bool)
		for _, host := range cs.Hosts {
			csHostSet[host] = true
		}

		optionHostSet := make(map[string]bool)
		for _, host := range seedlist {
			optionHostSet[host] = true
		}

		// check the sets are equal
		if len(csHostSet) != len(optionHostSet) {
			return ConflictingArgsErrorFormat("host", strings.Join(cs.Hosts, ","), opts.Host, command.MongoDBHost)
		}

		for host := range csHostSet {
			if _, ok := optionHostSet[host]; !ok {
				return ConflictingArgsErrorFormat("host", strings.Join(cs.Hosts, ","), opts.Host, command.MongoDBHost)
			}
		}
	} else if len(cs.Hosts) > 0 {
		if cs.ReplicaSet != "" {
			opts.Host = cs.ReplicaSet + "/"
		}

		// check if there is a <host:port> pair with a port that matches --port <port>
		conflictingPorts := true
		for _, host := range cs.Hosts {
			hostPort := strings.Split(host, ":")
			opts.Host += hostPort[0] + ","

			// a port might not be specified, e.g. `mongostat --discover`
			if len(hostPort) == 2 {
				if opts.Port != "" {
					if hostPort[1] == opts.Port {
						conflictingPorts = false
					}
				} else {
					opts.Port = hostPort[1]
					conflictingPorts = false
				}
			} else {
				conflictingPorts = false
			}
		}
		if conflictingPorts {
			return ConflictingArgsErrorFormat("port", strings.Join(cs.Hosts, ","), opts.Port, command.MongoDBPort)
		}
		// remove trailing comma
		opts.Host = opts.Host[:len(opts.Host)-1]
	}

	if len(cs.Hosts) > 1 && cs.LoadBalanced {
		return fmt.Errorf("loadBalanced cannot be set to true if multiple hosts are specified")
	}

	if opts.Connection.ServerSelectionTimeout != 0 && cs.ServerSelectionTimeoutSet {
		if (time.Duration(opts.Connection.ServerSelectionTimeout) * time.Millisecond) != cs.ServerSelectionTimeout {
			return ConflictingArgsErrorFormat("serverSelectionTimeout", strconv.Itoa(int(cs.ServerSelectionTimeout/time.Millisecond)), strconv.Itoa(opts.Connection.ServerSelectionTimeout), "--serverSelectionTimeout")
		}
	}
	if opts.Connection.ServerSelectionTimeout != 0 && !cs.ServerSelectionTimeoutSet {
		cs.ServerSelectionTimeout = time.Duration(opts.Connection.ServerSelectionTimeout) * time.Millisecond
		cs.ServerSelectionTimeoutSet = true
	}
	if opts.Connection.ServerSelectionTimeout == 0 && cs.ServerSelectionTimeoutSet {
		opts.Connection.ServerSelectionTimeout = int(cs.ServerSelectionTimeout / time.Millisecond)
	}

	if opts.Connection.Timeout != 3 && cs.ConnectTimeoutSet {
		if (time.Duration(opts.Connection.Timeout) * time.Millisecond) != cs.ConnectTimeout {
			return ConflictingArgsErrorFormat("connectTimeout", strconv.Itoa(int(cs.ConnectTimeout/time.Millisecond)), strconv.Itoa(opts.Connection.Timeout), "--dialTimeout")
		}
	}
	if opts.Connection.Timeout != 3 && !cs.ConnectTimeoutSet {
		cs.ConnectTimeout = time.Duration(opts.Connection.Timeout) * time.Millisecond
		cs.ConnectTimeoutSet = true
	}
	if opts.Connection.Timeout == 3 && cs.ConnectTimeoutSet {
		opts.Connection.Timeout = int(cs.ConnectTimeout / time.Millisecond)
	}

	if opts.Connection.SocketTimeout != 0 && cs.SocketTimeoutSet {
		if (time.Duration(opts.Connection.SocketTimeout) * time.Millisecond) != cs.SocketTimeout {
			return ConflictingArgsErrorFormat("SocketTimeout", strconv.Itoa(int(cs.SocketTimeout/time.Millisecond)), strconv.Itoa(opts.Connection.SocketTimeout), "--socketTimeout")
		}
	}
	if opts.Connection.SocketTimeout != 0 && !cs.SocketTimeoutSet {
		cs.SocketTimeout = time.Duration(opts.Connection.SocketTimeout) * time.Millisecond
		cs.SocketTimeoutSet = true
	}
	if opts.Connection.SocketTimeout == 0 && cs.SocketTimeoutSet {
		opts.Connection.SocketTimeout = int(cs.SocketTimeout / time.Millisecond)
	}

	if len(cs.Compressors) != 0 {
		if opts.Connection.Compressors != "none" && opts.Connection.Compressors != strings.Join(cs.Compressors, ",") {
			return ConflictingArgsErrorFormat("compressors", strings.Join(cs.Compressors, ","), opts.Connection.Compressors, "--compressors")
		}
	} else {
		cs.Compressors = strings.Split(opts.Connection.Compressors, ",")
	}

	if opts.Username != "" && cs.Username != "" {
		if opts.Username != cs.Username {
			return ConflictingArgsErrorFormat("username", cs.Username, opts.Username, command.MongoDBUsername)
		}
	}
	if opts.Username != "" && cs.Username == "" {
		cs.Username = opts.Username
	}
	if opts.Username == "" && cs.Username != "" {
		opts.Username = cs.Username
	}

	if opts.Password != "" && cs.PasswordSet {
		if opts.Password != cs.Password {
			return fmt.Errorf("invalid Options: Cannot specify different password in connection URI and command-line option")
		}
	}
	if opts.Password != "" && !cs.PasswordSet {
		cs.Password = opts.Password
		cs.PasswordSet = true
	}
	if opts.Password == "" && cs.PasswordSet {
		opts.Password = cs.Password
	}

	if opts.Source != "" && cs.AuthSourceSet {
		if opts.Source != cs.AuthSource {
			return ConflictingArgsErrorFormat("authSource", cs.AuthSource, opts.Source, command.MongoDBAuthDatabase)
		}
	}
	if opts.Source != "" && !cs.AuthSourceSet {
		cs.AuthSource = opts.Source
		cs.AuthSourceSet = true
	}
	if opts.Source == "" && cs.AuthSourceSet {
		opts.Source = cs.AuthSource
	}

	if opts.Mechanism != "" && cs.AuthMechanism != "" {
		if opts.Mechanism != cs.AuthMechanism {
			return ConflictingArgsErrorFormat("authMechanism", cs.AuthMechanism, opts.Mechanism, command.MongoDBAuthMechanism)
		}
	}
	if opts.Mechanism != "" && cs.AuthMechanism == "" {
		cs.AuthMechanism = opts.Mechanism
	}
	if opts.Mechanism == "" && cs.AuthMechanism != "" {
		opts.Mechanism = cs.AuthMechanism
	}

	if opts.DB != "" && cs.Database != "" {
		if opts.DB != cs.Database {
			return ConflictingArgsErrorFormat("database", cs.Database, opts.DB, command.MongoDBDatabase)
		}
	}
	if opts.DB != "" && cs.Database == "" {
		cs.Database = opts.DB
	}
	if opts.DB == "" && cs.Database != "" {
		opts.DB = cs.Database
	}

	// check replica set name equality
	if opts.ReplicaSetName != "" && cs.ReplicaSet != "" {
		if opts.ReplicaSetName != cs.ReplicaSet {
			return ConflictingArgsErrorFormat("replica set name", cs.ReplicaSet, opts.Host, command.MongoDBHost)
		}
		if opts.ConnString.LoadBalanced {
			return fmt.Errorf("loadBalanced cannot be set to true if the replica set name is specified")
		}
	}
	if opts.ReplicaSetName != "" && cs.ReplicaSet == "" {
		cs.ReplicaSet = opts.ReplicaSetName
	}
	if opts.ReplicaSetName == "" && cs.ReplicaSet != "" {
		opts.ReplicaSetName = cs.ReplicaSet
	}

	// Connect directly to a host if indicated by the connection string.
	opts.Direct = cs.DirectConnection || (cs.Connect == connstring.SingleConnect)
	if opts.Direct && opts.ConnString.LoadBalanced {
		return fmt.Errorf("loadBalanced cannot be set to true if the direct connection option is specified")
	}

	if cs.RetryWritesSet {
		opts.RetryWrites = &cs.RetryWrites
	}

	if cs.SSLSet {
		if opts.UseSSL && !cs.SSL {
			return ConflictingArgsErrorFormat("ssl", strconv.FormatBool(cs.SSL), strconv.FormatBool(opts.UseSSL), command.MongoDBSSL)
		} else if !opts.UseSSL && cs.SSL {
			opts.UseSSL = cs.SSL
		}
	}

	// ignore opts.UseSSL being false due to zero-value problem (TOOLS-2459 PR for details)
	// Ignore: opts.UseSSL = false, cs.SSL = true (have cs take precedence)
	// Treat as conflict: opts.UseSSL = true, cs.SSL = false
	if opts.UseSSL && cs.SSLSet {
		if !cs.SSL {
			return ConflictingArgsErrorFormat("ssl or tls", strconv.FormatBool(cs.SSL), strconv.FormatBool(opts.UseSSL), command.MongoDBSSL)
		}
	}
	if opts.UseSSL && !cs.SSLSet {
		cs.SSL = opts.UseSSL
		cs.SSLSet = true
	}
	// If SSL set in cs but not in opts,
	if !opts.UseSSL && cs.SSLSet {
		opts.UseSSL = cs.SSL
	}

	if opts.SSLCAFile != "" && cs.SSLCaFileSet {
		if opts.SSLCAFile != cs.SSLCaFile {
			return ConflictingArgsErrorFormat("sslCAFile", cs.SSLCaFile, opts.SSLCAFile, command.MongoDBSSLCAFile)
		}
	}
	if opts.SSLCAFile != "" && !cs.SSLCaFileSet {
		cs.SSLCaFile = opts.SSLCAFile
		cs.SSLCaFileSet = true
	}
	if opts.SSLCAFile == "" && cs.SSLCaFileSet {
		opts.SSLCAFile = cs.SSLCaFile
	}

	if opts.SSLPEMKeyFile != "" && cs.SSLClientCertificateKeyFileSet {
		if opts.SSLPEMKeyFile != cs.SSLClientCertificateKeyFile {
			return ConflictingArgsErrorFormat("sslClientCertificateKeyFile", cs.SSLClientCertificateKeyFile, opts.SSLPEMKeyFile, command.MongoDBSSLPEMKeyFile)
		}
	}
	if opts.SSLPEMKeyFile != "" && !cs.SSLClientCertificateKeyFileSet {
		cs.SSLClientCertificateKeyFile = opts.SSLPEMKeyFile
		cs.SSLClientCertificateKeyFileSet = true
	}
	if opts.SSLPEMKeyFile == "" && cs.SSLClientCertificateKeyFileSet {
		opts.SSLPEMKeyFile = cs.SSLClientCertificateKeyFile
	}

	if opts.SSLPEMKeyPassword != "" && cs.SSLClientCertificateKeyPasswordSet {
		if opts.SSLPEMKeyPassword != cs.SSLClientCertificateKeyPassword() {
			return ConflictingArgsErrorFormat("sslPEMKeyFilePassword", cs.SSLClientCertificateKeyPassword(), opts.SSLPEMKeyPassword, command.MongoDBSSLPEMKeyPassword)
		}
	}
	if opts.SSLPEMKeyPassword != "" && !cs.SSLClientCertificateKeyPasswordSet {
		cs.SSLClientCertificateKeyPassword = func() string { return opts.SSLPEMKeyPassword }
		cs.SSLClientCertificateKeyPasswordSet = true
	}
	if opts.SSLPEMKeyPassword == "" && cs.SSLClientCertificateKeyPasswordSet {
		opts.SSLPEMKeyPassword = cs.SSLClientCertificateKeyPassword()
	}

	// Note: SSLCRLFile is not parsed by the go driver

	// ignore (opts.SSLAllowInvalidCert || opts.SSLAllowInvalidHost) being false due to zero-value problem (TOOLS-2459 PR for details)
	// Have cs take precedence in cases where it is unclear
	if (opts.SSLAllowInvalidCert || opts.SSLAllowInvalidHost || opts.TLSInsecure) && cs.SSLInsecureSet {
		if !cs.SSLInsecure {
			return ConflictingArgsErrorFormat("sslInsecure or tlsInsecure", "false", "true", "--sslAllowInvalidCert or --sslAllowInvalidHost")
		}
	}
	if (opts.SSLAllowInvalidCert || opts.SSLAllowInvalidHost || opts.TLSInsecure) && !cs.SSLInsecureSet {
		cs.SSLInsecure = true
		cs.SSLInsecureSet = true
	}
	if (!opts.SSLAllowInvalidCert && !opts.SSLAllowInvalidHost || !opts.TLSInsecure) && cs.SSLInsecureSet {
		opts.SSLAllowInvalidCert = cs.SSLInsecure
		opts.SSLAllowInvalidHost = cs.SSLInsecure
		opts.TLSInsecure = cs.SSLInsecure
	}

	if strings.ToLower(cs.AuthMechanism) == "gssapi" {

		gssapiServiceName, _ := cs.AuthMechanismProperties["SERVICE_NAME"]

		if opts.Kerberos.Service != "" && cs.AuthMechanismPropertiesSet {
			if opts.Kerberos.Service != gssapiServiceName {
				return ConflictingArgsErrorFormat("Kerberos service name", gssapiServiceName, opts.Kerberos.Service, command.MongoDBGSSAPIServiceName)
			}
		}
		if opts.Kerberos.Service != "" && !cs.AuthMechanismPropertiesSet {
			if cs.AuthMechanismProperties == nil {
				cs.AuthMechanismProperties = make(map[string]string)
			}
			cs.AuthMechanismProperties["SERVICE_NAME"] = opts.Kerberos.Service
			cs.AuthMechanismPropertiesSet = true
		}
		if opts.Kerberos.Service == "" && cs.AuthMechanismPropertiesSet {
			opts.Kerberos.Service = gssapiServiceName
		}
	}

	if strings.ToLower(cs.AuthMechanism) == "mongodb-aws" {
		awsSessionToken, _ := cs.AuthMechanismProperties["AWS_SESSION_TOKEN"]

		if opts.AWSSessionToken != "" && cs.AuthMechanismPropertiesSet {
			if opts.AWSSessionToken != awsSessionToken {
				return ConflictingArgsErrorFormat("AWS Session Token", awsSessionToken, opts.AWSSessionToken, command.MongoDBAWSSessionToken)
			}
		}
		if opts.AWSSessionToken != "" && !cs.AuthMechanismPropertiesSet {
			if cs.AuthMechanismProperties == nil {
				cs.AuthMechanismProperties = make(map[string]string)
			}
			cs.AuthMechanismProperties["AWS_SESSION_TOKEN"] = opts.AWSSessionToken
			cs.AuthMechanismPropertiesSet = true
		}
		if opts.AWSSessionToken == "" && cs.AuthMechanismPropertiesSet {
			opts.AWSSessionToken = awsSessionToken
		}
	}

	// set the connString on opts so it can be validated later
	opts.ConnString = cs

	return nil
}

// GetAuthenticationDatabase Get the authentication database to use. Should be the value of
// --authenticationDatabase if it's provided, otherwise, the database that's
// specified in the tool's --db arg.
func (opts *Options) GetAuthenticationDatabase() string {
	if opts.Auth.Source != "" {
		return opts.Auth.Source
	} else if opts.Auth.RequiresExternalDB() {
		return "$external"
	} else if opts.Namespace != nil && opts.Namespace.DB != "" {
		return opts.Namespace.DB
	}
	return ""
}

// General Struct holding generic options
type General struct {
	Failpoints string
	Trace      bool
}

type URI struct {
	ConnectionString string
	ConnString       connstring.ConnString
}

func NewURI(unparsed string) (*URI, error) {
	cs, err := connstring.Parse(unparsed)
	if err != nil {
		return nil, fmt.Errorf("error parsing URI from %v: %v", unparsed, err)
	}
	return &URI{ConnectionString: cs.String(), ConnString: *cs}, nil
}

func (uri *URI) GetConnectionAddrs() []string {
	return uri.ConnString.Hosts
}
func (uri *URI) ParsedConnString() *connstring.ConnString {
	if uri.ConnectionString == "" {
		return nil
	}
	return &uri.ConnString
}

// Connection Struct holding connection-related options
type Connection struct {
	Host string
	Port string

	Timeout                int
	SocketTimeout          int
	TCPKeepAliveSeconds    int
	ServerSelectionTimeout int
	Compressors            string
}

// SSL Struct holding ssl-related options
type SSL struct {
	UseSSL              bool
	SSLCAFile           string
	SSLPEMKeyFile       string
	SSLPEMKeyPassword   string
	SSLCRLFile          string
	SSLAllowInvalidCert bool
	SSLAllowInvalidHost bool
	SSLFipsMode         bool
	TLSInsecure         bool
}

// ShouldAskForPassword returns true if the user specifies a ssl pem key file
// flag but no password for that file, and the key file has any encrypted
// blocks.
func (ssl *SSL) ShouldAskForPassword() (bool, error) {
	if ssl.SSLPEMKeyFile == "" || ssl.SSLPEMKeyPassword != "" {
		return false, nil
	}
	return ssl.pemKeyFileHasEncryptedKey()
}

func (ssl *SSL) pemKeyFileHasEncryptedKey() (bool, error) {
	b, err := os.ReadFile(ssl.SSLPEMKeyFile)
	if err != nil {
		return false, err
	}

	for {
		var v *pem.Block
		v, b = pem.Decode(b)
		if v == nil {
			break
		}
		if v.Type == "ENCRYPTED PRIVATE KEY" {
			return true, nil
		}
	}

	return false, nil
}

// Auth Struct holding auth-related options
type Auth struct {
	Username        string
	Password        string
	Source          string
	Mechanism       string
	AWSSessionToken string
}

func (auth *Auth) RequiresExternalDB() bool {
	return auth.Mechanism == "GSSAPI" || auth.Mechanism == "PLAIN" || auth.Mechanism == "MONGODB-X509"
}

func (auth *Auth) IsSet() bool {
	return *auth != Auth{}
}

// ShouldAskForPassword returns true if the user specifies a username flag
// but no password, and the authentication mechanism requires a password.
func (auth *Auth) ShouldAskForPassword() bool {
	return auth.Username != "" && auth.Password == "" &&
		!(auth.Mechanism == "MONGODB-X509" || auth.Mechanism == "GSSAPI")
}

// Kerberos Struct for Kerberos/GSSAPI-specific options
type Kerberos struct {
	Service     string
	ServiceHost string
}

type Namespace struct {
	// Specified database and collection
	DB         string `short:"d" long:"db" value-name:"<database-name>" description:"database to use"`
	Collection string `short:"c" long:"collection" value-name:"<collection-name>" description:"collection to use"`
}

func (ns Namespace) String() string {
	return ns.DB + "." + ns.Collection
}

type WriteConcern struct {
	// Specifies the write concern for each write operation that mongofiles writes to the target database.
	// By default, mongofiles waits for a majority of members from the replica set to respond before returning.
	WriteConcern string

	w        int
	wtimeout int
	fsync    bool
	journal  bool
}

// QueryOptions defines the set of options to use in retrieving data from the server.
type QueryOptions struct {
	Query          string
	QueryFile      string
	SlaveOk        bool
	ReadPreference string
	ForceTableScan bool
	Skip           int64
	Limit          int64
	Sort           string
	AssertExists   bool
}

func ConflictingArgsErrorFormat(optionName, uriValue, cliValue, cliOptionName string) error {
	return fmt.Errorf("invalid Options: cannot specify different %s in connection URI and option (\"%s\" was specified in the URI and \"%s\" was specified in the --%s option)", optionName, uriValue, cliValue, cliOptionName)
}
