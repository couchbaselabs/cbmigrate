package dynamodb

import (
	"github.com/couchbaselabs/cbmigrate/internal/dynamodb"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/dynamodb/command"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase"
	cRepo "github.com/couchbaselabs/cbmigrate/internal/couchbase/repo"
	dOpts "github.com/couchbaselabs/cbmigrate/internal/dynamodb/option"
	dRepo "github.com/couchbaselabs/cbmigrate/internal/dynamodb/repo"
	"github.com/couchbaselabs/cbmigrate/internal/migrater"
)

type Action struct {
	Migrate migrater.IMigrate[dOpts.Options]
}

func NewAction() *Action {
	return &Action{
		Migrate: migrater.NewMigrator(
			dynamodb.NewDynamoDB(dRepo.NewRepo()),
			couchbase.NewCouchbase(cRepo.NewRepo()),
		),
	}
}

func (a *Action) RunE(cmd *cobra.Command, args []string) error {

	var missingRequiredOptions []string
	if !cmd.Flags().Changed(command.DynamoDBTableName) {
		missingRequiredOptions = append(missingRequiredOptions, command.DynamoDBTableName)
	}
	missingRequiredOptions = append(missingRequiredOptions, common.CouchBaseMissingRequiredOptions(cmd)...)
	if len(missingRequiredOptions) > 0 {
		err := common.ReqFieldsError(missingRequiredOptions)
		if err != nil {
			return err
		}
	}
	dopts := &dOpts.Options{}

	dopts.TableName, _ = cmd.Flags().GetString(command.DynamoDBTableName)
	dopts.EndpointUrl, _ = cmd.Flags().GetString(command.DynamoDBEndpointURL)
	dopts.Profile, _ = cmd.Flags().GetString(command.DynamoDBProfile)
	dopts.Region, _ = cmd.Flags().GetString(command.DynamoDBRegion)
	dopts.CABundle, _ = cmd.Flags().GetString(command.DynamoDBCaBundle)
	insecure, _ := cmd.Flags().GetBool(command.DynamoDBNoVerifySSL)
	dopts.NoSSLVerify = insecure

	cbOpts, err := common.ParesCouchbaseOptions(cmd, dopts.TableName)
	if err != nil {
		return err
	}
	copyIndexes, _ := cmd.Flags().GetBool(common.CopyIndexes)
	bufferSize, _ := cmd.Flags().GetInt(common.BufferSize)
	err = a.Migrate.Copy(dopts, cbOpts, copyIndexes, bufferSize)
	if err != nil {
		zap.S().Fatal(err)
	}
	return nil
}

func GetDynamoDBMigrateCommand() *cobra.Command {
	cmd := command.NewCommand()
	action := NewAction()
	cmd.RunE = action.RunE
	return cmd
}
