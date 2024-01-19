package common

import (
	"bytes"
	"fmt"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// ExecuteCommand function to run the test cases
func ExecuteCommand(cmd *cobra.Command, args ...string) (output string, err error) {
	// Use a bytes buffer to capture the command's output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err = cmd.Execute()

	return buf.String(), err
}

func ValidateMustAllOrNotFlag(cmd *cobra.Command, flags ...string) error {
	var missing []string
	for _, v := range flags {
		if !cmd.Flags().Changed(v) {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 && len(missing) != len(flags) {
		return fmt.Errorf("inconsistent flag usage. Flags %s must all be provided together or not at all. Missing: %s", strings.Join(flags, ", "), strings.Join(missing, ", "))
	}
	return nil
}

func CopyMap[T1 comparable, T2 any](m map[T1]T2) map[T1]T2 {
	newMap := make(map[T1]T2)
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func ReqFieldsValidation(cmd *cobra.Command, flags []string) error {
	var missingFields []string
	for _, flag := range flags {
		if !cmd.Flags().Changed(flag) {
			missingFields = append(missingFields, flag)
		}
	}
	if len(missingFields) > 0 {
		return ReqFieldsError(missingFields)
	}
	return nil
}
func ReqFieldsError(missingFields []string) error {
	return fmt.Errorf("required flag(s) \"%s\" not set", strings.Join(missingFields, "\", \""))
}

// Extract the replica set name and the list of hosts from the connection string
func SplitHostArg(connString string) ([]string, string) {

	// strip off the replica set name from the beginning
	slashIndex := strings.Index(connString, "/")
	setName := ""
	if slashIndex != -1 {
		setName = connString[:slashIndex]
		if slashIndex == len(connString)-1 {
			return []string{""}, setName
		}
		connString = connString[slashIndex+1:]
	}

	// split the hosts, and return them and the set name
	return strings.Split(connString, ","), setName
}

// BuildURI assembles a URI from host and port arguments, including a possible
// replica set name on the host part
func BuildURI(host, port string) string {
	seedlist, setname := SplitHostArg(host)

	// if any seedlist entry is empty, make it localhost
	for i := range seedlist {
		if seedlist[i] == "" {
			seedlist[i] = "localhost"
		}
	}

	// if a port is provided, append it to any host without a port; if any
	// host part is empty string, make it localhost
	if port != "" {
		for i := range seedlist {
			if strings.Index(seedlist[i], ":") == -1 {
				seedlist[i] = seedlist[i] + ":" + port
			}
		}
	}

	hostpairs := strings.Join(seedlist, ",")
	if setname != "" {
		return fmt.Sprintf("mongodb://%s/?replicaSet=%s", hostpairs, setname)
	}
	return fmt.Sprintf("mongodb://%s/", hostpairs)
}

func ParesCouchbaseOptions(cmd *cobra.Command) (*option.Options, error) {
	var err error
	cbopts := &option.Options{
		Auth:      &option.Auth{},
		SSL:       &option.SSL{},
		NameSpace: &option.NameSpace{},
	}
	cbopts.Cluster, _ = cmd.Flags().GetString(CBCluster)
	cbopts.Auth.Username, _ = cmd.Flags().GetString(CBUsername)
	cbopts.Auth.Password, _ = cmd.Flags().GetString(CBPassword)
	cbClientCert, _ := cmd.Flags().GetString(CBClientCert)
	if cbClientCert != "" {
		cbopts.Auth.ClientCert, err = os.ReadFile(cbClientCert)
		if err != nil {
			return nil, err
		}
	}
	cbopts.Auth.ClientCertPassword, _ = cmd.Flags().GetString(CBClientCertPassword)

	cbClientKey, _ := cmd.Flags().GetString(CBClientKey)
	if cbClientKey != "" {
		cbopts.Auth.ClientKey, err = os.ReadFile(cbClientKey)
		if err != nil {
			return nil, err
		}
	}
	cbopts.Auth.ClientKeyPassword, _ = cmd.Flags().GetString(CBClientKeyPassword)

	cbCACert, _ := cmd.Flags().GetString(CBCACert)
	if cbCACert != "" {
		cbopts.SSL.CaCert, err = os.ReadFile(cbCACert)
		if err != nil {
			return nil, err
		}
	}
	cbopts.SSL.NoSSLVerify, _ = cmd.Flags().GetBool(CBNoSSLVerify)

	cbopts.NameSpace.Bucket, _ = cmd.Flags().GetString(CBBucket)
	cbopts.NameSpace.Scope, _ = cmd.Flags().GetString(CBScope)
	cbopts.NameSpace.Collection, _ = cmd.Flags().GetString(CBCollection)

	cbopts.GeneratedKey, _ = cmd.Flags().GetString(CBGenerateKey)

	cbopts.BatchSize = 250
	return cbopts, nil
}

func CouchBaseMissingRequiredOptions(cmd *cobra.Command) []string {
	var missingRequiredOptions []string
	switch {
	case !cmd.Flags().Changed(CBCluster):
		missingRequiredOptions = append(missingRequiredOptions, CBCluster)
		fallthrough
	case !cmd.Flags().Changed(CBScope):
		missingRequiredOptions = append(missingRequiredOptions, CBScope)
		fallthrough
	case !cmd.Flags().Changed(CBCollection):
		missingRequiredOptions = append(missingRequiredOptions, CBCollection)
	}
	return missingRequiredOptions
}
