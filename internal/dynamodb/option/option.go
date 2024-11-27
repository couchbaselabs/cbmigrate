package option

type Options struct {
	TableName   string
	EndpointUrl string
	NoSSLVerify bool
	Profile     string
	AccessKey   string
	SecretKey   string
	Region      string
	CABundle    string
	segments    int
	limit       int
}
