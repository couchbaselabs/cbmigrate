package dynamodb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsHTTP "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/couchbaselabs/cbmigrate/internal/dynamodb/option"
	"net/http"
	"os"
)

type DB struct {
	*dynamodb.Client
}

func (d *DB) Init(opts *option.Options) error {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files

	var awsOpts []func(*config.LoadOptions) error
	switch {
	case opts.Region != "":
		awsOpts = append(awsOpts, config.WithRegion(opts.Region))
		fallthrough
	case opts.Profile != "":
		awsOpts = append(awsOpts, config.WithSharedConfigProfile(opts.Profile))
	case opts.CABundle != "":
		caCert, err := os.ReadFile(opts.CABundle)
		if err != nil {
			panic(fmt.Errorf("unable to load CA bundle: %w", err))
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		httpClient := awsHTTP.NewBuildableClient().WithTransportOptions(func(transport *http.Transport) {
			transport.TLSClientConfig = &tls.Config{
				RootCAs: caCertPool, // Set RootCAs to your custom CA pool
			}
		})
		awsOpts = append(awsOpts, config.WithHTTPClient(httpClient))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), awsOpts...)
	if err != nil {
		return fmt.Errorf("err: %w", err)
	}
	// Using the Config value, create the DynamoDB client
	d.Client = dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {

		switch {
		case opts.EndpointUrl != "":
			options.BaseEndpoint = aws.String(opts.EndpointUrl)
			fallthrough
		case opts.NoSSLVerify:
			options.EndpointOptions.DisableHTTPS = true
		}

	})
	return nil
}
