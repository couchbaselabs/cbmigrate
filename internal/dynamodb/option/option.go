package option

type Options struct {
	TableName   string
	EndpointUrl string
	NoSSLVerify bool
	Profile     string
	Region      string
	CABundle    string
}
