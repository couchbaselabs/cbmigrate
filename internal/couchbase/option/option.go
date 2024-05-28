package option

type Options struct {
	Cluster string
	*Auth
	*SSL
	*NameSpace
	GeneratedKey    string
	HashDocumentKey string
	BatchSize       int
}

type Auth struct {
	Username           string
	Password           string
	ClientCert         []byte
	ClientCertPassword string
	ClientKey          []byte
	ClientKeyPassword  string
}

type SSL struct {
	CaCert      []byte
	NoSSLVerify bool
}

type NameSpace struct {
	Bucket     string
	Scope      string
	Collection string
}
