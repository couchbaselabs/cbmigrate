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
	DynamoDBSegments    = "dynamodb-segments"
	DynamoDBLimit       = "dynamodb-limit"
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
	Name:     DynamoDBAccessKey,
	Usage:    "AWS Access Key ID.",
	Required: true,
}

var dynamoDBSecretKey = &flag.StringFlag{
	Name:     DynamoDBSecretKey,
	Usage:    "AWS Secret Access Key.",
	Required: true,
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

var dynamoDBSegments = &flag.IntFlag{
	Name:  DynamoDBSegments,
	Value: 1,
	Usage: "Specifies the total number of segments to divide the DynamoDB table into for parallel scanning. Each segment is scanned independently, " +
		"allowing multiple threads or processes to work concurrently for faster data retrieval. Use this option to optimize performance for large tables." +
		"By default entire table is scanned sequentially without segmentation",
}

var dynamoDBLimit = &flag.IntFlag{
	Name: DynamoDBLimit,
	Usage: "Specifies the maximum number of items to retrieve per page during a scan operation. " +
		"Use this option to control the amount of data fetched in a single request, " +
		"helping to manage memory usage and API call rates during scanning.",
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
			Type: flag.RelationshipOR,
		},
		dynamoDBRegion,
		dynamoDBEndpointURL,
		dynamoDBNoVerifySSL,
		dynamoDBCaBundle,
		dynamoDBSegments,
		dynamoDBLimit,
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
			Value: "cbmigrate dynamodb --dynamodb-table-name da-test-2 --aws-access-key-id aws-access-key-id --aws-secret-access-key aws-secret-access-key --aws-region aws-region --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name",
			Usage: "With aws access key id and aws secret access key options.",
		},
		{
			Value: "cbmigrate dynamodb --dynamodb-table-name da-test-2 --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::%firstname%::%lastname% --hash-document-key sha256",
			Usage: "With hash document key option.",
		},
	}
	usage := "Migrate data from DynamoDB to Couchbase"
	return common.NewCommand(common.DynamoDB, []string{"d"}, examples, usage, usage, flags)
}
