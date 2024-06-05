package command

import (
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/flag"
	"github.com/spf13/cobra"
)

const (
	DynamoDBEndpointURL = "aws-endpoint-url"
	DynamoDBNoVerifySSL = "aws-no-verify-ssl"
	DynamoDBProfile     = "aws-profile"
	DynamoDBAccessKey   = "aws-access-key-id"
	DynamoDBSecretKey   = "aws-secret-access-key"
	DynamoDBRegion      = "aws-region"
	DynamoDBCaBundle    = "aws-ca-bundle"
	DynamoDBTableName   = "dynamodb-table-name"
)

var dynamoDBEndpointURL = &flag.StringFlag{
	Name:  DynamoDBEndpointURL,
	Usage: "Override aws’s default default endpoint url with the given URL.",
}

var dynamoDBNoVerifySSL = &flag.BoolFlag{
	Name:  DynamoDBNoVerifySSL,
	Usage: "By default, the CLI uses SSL when communicating with AWS services. For each SSL connection, the CLI will verify SSL certificates. This option overrides the default behavior of verifying SSL certificates.",
}

var dynamoDBProfile = &flag.StringFlag{
	Name:  DynamoDBProfile,
	Usage: "Use a specific aws profile from your credential file.",
}

var dynamoDBAccessKey = &flag.StringFlag{
	Name:  DynamoDBAccessKey,
	Usage: "AWS Access Key.",
}

var dynamoDBSecretKey = &flag.StringFlag{
	Name:  DynamoDBSecretKey,
	Usage: "AWS Secret Key.",
}

var dynamoDBRegion = &flag.StringFlag{
	Name:  DynamoDBRegion,
	Usage: "The region to use. Overrides config/env settings.",
}

var dynamoDBCaBundle = &flag.StringFlag{
	Name:  DynamoDBCaBundle,
	Usage: "The CA certificate bundle to use when verifying SSL certificates. Overrides config/env settings.",
}

var dynamoDBTableName = &flag.StringFlag{
	Name:     DynamoDBTableName,
	Usage:    "The name of the table containing the requested item. You can also provide the Amazon Resource Name (ARN) of the table in this parameter.",
	Required: true,
}

func NewCommand() *cobra.Command {

	//short := "A tool to convert time series data in CSV to the one supported by Couchbase."
	//long := `cbmigrate dynamodb is a CLI tool for Couchbase that enables users to convert time series data in CSV to the format required by Couchbase.`
	flags := []flag.Flag{
		dynamoDBTableName,
		&flag.CompositeFlag{
			Flags: []flag.Flag{
				dynamoDBProfile,
				&flag.CompositeFlag{
					Flags: []flag.Flag{
						dynamoDBAccessKey,
						dynamoDBSecretKey,
					},
				},
			},
		},
		dynamoDBRegion,
		dynamoDBEndpointURL,
		dynamoDBNoVerifySSL,
		dynamoDBCaBundle,
	}
	flags = append(flags, common.GetCBFlags()...)
	flags = append(flags, common.GetCBGenerateKeyOption(""))
	flags = append(flags, common.GetCommonFlags()...)
	examples := []common.Example{
		{
			Value: "cbmigrate dynamodb --dynamodb-table-name da-test-2 --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name",
			Usage: "Imports data from DynamoDB to Couchbase.",
		},
		{
			Value: "cbmigrate dynamodb --dynamodb-table-name da-test-2 --aws-profile aws-profile --aws-region aws-region --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::#UUID#",
			Usage: "With aws profile and region and couchbase collection and generator key options.",
		},
		{
			Value: "cbmigrate dynamodb --dynamodb-table-name da-test-2 --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::%firstname%::%lastname% --hash-document-key sha256",
			Usage: "With hash document key option.",
		},
	}
	usage := "Migrate data from DynamoDB to Couchbase"
	return common.NewCommand(common.DynamoDB, []string{"d"}, examples, usage, usage, flags)
}
