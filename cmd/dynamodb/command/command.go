package command

import (
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/flag"
	"github.com/spf13/cobra"
)

const (
	DynamoDBEndpointURL = "dynamodb-endpoint-url"
	DynamoDBNoVerifySSL = "dynamodb-no-verify-ssl"
	DynamoDBProfile     = "dynamodb-profile"
	DynamoDBRegion      = "dynamodb-region"
	DynamoDBCaBundle    = "ca-bundle"
	DynamoDBTableName   = "dynamodb-table-name"
)

var dynamoDBEndpointURL = &flag.StringFlag{
	Name:     DynamoDBEndpointURL,
	Usage:    "Override command’s default URL with the given URL",
	Required: false,
}

var dynamoDBNoVerifySSL = &flag.BoolFlag{
	Name:     DynamoDBNoVerifySSL,
	Usage:    "By default, the CLI uses SSL when communicating with AWS services. For each SSL connection, the CLI will verify SSL certificates. This option overrides the default behavior of verifying SSL certificates",
	Required: true,
}

var dynamoDBProfile = &flag.StringFlag{
	Name:  DynamoDBProfile,
	Usage: "Use a specific aws profile from your credential filed",
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
	Name:  DynamoDBTableName,
	Usage: "The name of the table containing the requested item. You can also provide the Amazon Resource Name (ARN) of the table in this parameter.",
}

func NewCommand() *cobra.Command {

	//short := "A tool to convert time series data in CSV to the one supported by Couchbase."
	//long := `cbmigrate dynamodb is a CLI tool for Couchbase that enables users to convert time series data in CSV to the format required by Couchbase.`
	flags := []flag.Flag{
		dynamoDBEndpointURL,
		dynamoDBNoVerifySSL,
		dynamoDBProfile,
		dynamoDBRegion,
		dynamoDBCaBundle,
		dynamoDBTableName,
	}
	flags = append(flags, common.GetCBFlags()...)
	flags = append(flags, common.GetCBGenerateKeyOption(""))
	flags = append(flags, common.GetCommonFlags()...)
	examples := []common.Example{
		{
			Value: "cbmigrate dynamodb --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name",
			Usage: "Imports data from DynamoDB to Couchbase.",
		},
	}
	usage := "Migrate data from DynamoDB to Couchbase"
	return common.NewCommand(common.DynamoDB, []string{"m"}, examples, usage, usage, flags)
}
