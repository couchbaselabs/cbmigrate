package couchbase

import (
	"crypto/tls"
	"github.com/couchbase/gocb/v2"
	tlsutil "github.com/couchbase/tools-common/http/tls"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
)

func createCouchbaseOptions(options *option.Options) (gocb.ClusterOptions, error) {
	cbOpts := gocb.ClusterOptions{}

	if options.Username != "" {
		auth := gocb.PasswordAuthenticator{
			Username: options.Username,
			Password: options.Password,
		}
		cbOpts.Authenticator = auth
	}
	if options.ClientCert != nil || options.CaCert != nil {
		// tls parsing code is similar to the code used in the cbimport.
		tlsConfig, err := tlsutil.NewConfig(tlsutil.ConfigOptions{
			ClientCert:     options.ClientCert,
			ClientKey:      options.ClientKey,
			Password:       []byte(getCertKeyPassword(options.ClientCertPassword, options.ClientKeyPassword)),
			ClientAuthType: tls.VerifyClientCertIfGiven,
			RootCAs:        options.CaCert,
			NoSSLVerify:    options.NoSSLVerify,
		})
		if err != nil {
			return gocb.ClusterOptions{}, err
		}

		if options.ClientCert != nil {
			auth := gocb.CertificateAuthenticator{
				ClientCertificate: &tlsConfig.Certificates[0],
			}
			cbOpts.Authenticator = auth
		}
		if options.CaCert != nil {
			cbOpts.SecurityConfig = gocb.SecurityConfig{
				TLSSkipVerify: options.NoSSLVerify == true,
				TLSRootCAs:    tlsConfig.RootCAs,
			}
		}
		if options.NoSSLVerify {
			cbOpts.SecurityConfig = gocb.SecurityConfig{
				TLSSkipVerify: options.NoSSLVerify == true,
			}
		}
	}
	return cbOpts, nil
}

// GetCertKeyPassword - Returns the password which should be used when creating a new TLS config.
func getCertKeyPassword(certPassword, keyPassword string) string {
	if keyPassword != "" {
		return keyPassword
	}

	return certPassword
}
