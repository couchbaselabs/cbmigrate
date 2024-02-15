package mongo

import (
	"github.com/spf13/cobra"

	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/mongo/command"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase"
	cRepo "github.com/couchbaselabs/cbmigrate/internal/couchbase/repo"
	"github.com/couchbaselabs/cbmigrate/internal/migrater"
	"github.com/couchbaselabs/cbmigrate/internal/mongo"
	mOpts "github.com/couchbaselabs/cbmigrate/internal/mongo/option"
	mRepo "github.com/couchbaselabs/cbmigrate/internal/mongo/repo"
	"github.com/couchbaselabs/cbmigrate/internal/option"
)

type Action struct {
	migrate migrater.IMigrate
}

func NewAction() *Action {
	return &Action{
		migrate: migrater.NewMigrator(
			mongo.NewMongo(mRepo.NewRepo()),
			couchbase.NewCouchbase(cRepo.NewRepo()),
			mongo.NewIndexFieldAnalyzer(),
		),
	}
}

func (a *Action) RunE(cmd *cobra.Command, args []string) error {

	var missingRequiredOpitons []string
	switch {
	case !cmd.Flags().Changed(command.MongoDBURI) && !cmd.Flags().Changed(command.MongoDBHost):
		missingRequiredOpitons = append(missingRequiredOpitons, command.MongoDBURI)
		fallthrough
	case !cmd.Flags().Changed(command.MongoDBCollection):
		missingRequiredOpitons = append(missingRequiredOpitons, command.MongoDBCollection)
	}
	missingRequiredOpitons = append(missingRequiredOpitons, common.CouchBaseMissingRequiredOptions(cmd)...)
	if len(missingRequiredOpitons) > 0 {
		err := common.ReqFieldsError(missingRequiredOpitons)
		if err != nil {
			return err
		}
	}
	mopts := &mOpts.Options{
		URI:        &mOpts.URI{},
		Connection: &mOpts.Connection{},
		SSL:        &mOpts.SSL{},
		Auth:       &mOpts.Auth{},
		Kerberos:   &mOpts.Kerberos{},
		Namespace:  &mOpts.Namespace{},
	}

	mopts.URI.ConnectionString, _ = cmd.Flags().GetString(command.MongoDBURI)
	mopts.Connection.Host, _ = cmd.Flags().GetString(command.MongoDBHost)
	mopts.Connection.Port, _ = cmd.Flags().GetString(command.MongoDBPort)

	insecure, _ := cmd.Flags().GetBool(command.MongoDBTLSInsecure)
	mopts.SSL.UseSSL = !insecure
	mopts.SSL.SSLCAFile, _ = cmd.Flags().GetString(command.MongoDBSSLCAFile)
	mopts.SSL.SSLPEMKeyFile, _ = cmd.Flags().GetString(command.MongoDBSSLPEMKeyFile)
	mopts.SSL.SSLPEMKeyPassword, _ = cmd.Flags().GetString(command.MongoDBSSLPEMKeyPassword)
	mopts.SSL.SSLFipsMode, _ = cmd.Flags().GetBool(command.MongoDBSSLFIPSMode)

	mopts.Auth.Username, _ = cmd.Flags().GetString(command.MongoDBUsername)
	mopts.Auth.Password, _ = cmd.Flags().GetString(command.MongoDBPassword)
	mopts.Auth.Source, _ = cmd.Flags().GetString(command.MongoDBAuthDatabase)
	mopts.Auth.Mechanism, _ = cmd.Flags().GetString(command.MongoDBAuthMechanism)
	mopts.Auth.AWSSessionToken, _ = cmd.Flags().GetString(command.MongoDBAWSSessionToken)

	mopts.Kerberos.ServiceHost, _ = cmd.Flags().GetString(command.MongoDBGSSAPIHostName)
	mopts.Kerberos.Service, _ = cmd.Flags().GetString(command.MongoDBGSSAPIServiceName)
	mopts.Namespace.DB, _ = cmd.Flags().GetString(command.MongoDBDatabase)
	mopts.Namespace.Collection, _ = cmd.Flags().GetString(command.MongoDBCollection)

	cbopts, err := common.ParesCouchbaseOptions(cmd)
	if err != nil {
		return err
	}
	if cbopts.GeneratedKey == "" {
		cbopts.GeneratedKey = " %_id%"
	}

	opts := &option.Options{
		MOpts:  mopts,
		CBOpts: cbopts,
	}

	err = a.migrate.Copy(opts)
	if err != nil {
		return err
	}
	return nil
}

func GetMongoMigrateCommand() *cobra.Command {
	cmd := command.NewCommand()
	action := NewAction()
	cmd.RunE = action.RunE
	return cmd
}
