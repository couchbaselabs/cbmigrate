
# Migrate Data from DynamoDB to Couchbase

This tool allows you to migrate data from a DynamoDB table to a Couchbase cluster.

## Features

- Direct migration from DynamoDb to Couchbase.
- SSL encryption support with optional verification.
- Customizable document key generation.
- Option to copy DynamoDb indexes.
- Debug output for detailed operation logs.

## Usage

```sh
cbmigrate dynamodb --dynamodb-table-name DYNAMODB_TABLE_NAME [[--aws-profile AWS_PROFILE] | [--aws-access-key-id AWS_ACCESS_KEY_ID --aws-secret-access-key AWS_SECRET_ACCESS_KEY]] [--aws-region AWS_REGION] [--aws-endpoint-url AWS_ENDPOINT_URL] [--aws-no-verify-ssl] [--aws-ca-bundle AWS_CA_BUNDLE] --cb-cluster CB_CLUSTER (--cb-username CB_USERNAME --cb-password CB_PASSWORD | --cb-client-cert CB_CLIENT_CERT [--cb-client-cert-password CB_CLIENT_CERT_PASSWORD] [--cb-client-key CB_CLIENT_KEY] [--cb-client-key-password CB_CLIENT_KEY_PASSWORD]) [--cb-cacert CB_CACERT] [--cb-no-ssl-verify] [--cb-bucket CB_BUCKET] [--cb-scope CB_SCOPE] [--cb-collection CB_COLLECTION] [--cb-batch-size CB_BATCH_SIZE] [--keep-primary-key] [--hash-document-key sha256,sha512] [--debug] [--cb-generate-key CB_GENERATE_KEY] [--copy-indexes] [--buffer-size BUFFER_SIZE] [--help HELP]
```

## Aliases

- `dynamodb`
- `d`

## Examples

- Imports data from DynamoDB to Couchbase.
```sh
cbmigrate dynamodb --dynamodb-table-name da-test-2 --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name
```

- With aws profile and region and couchbase collection and generator key options.
```sh
cbmigrate dynamodb --dynamodb-table-name da-test-2 --aws-profile aws-profile --aws-region aws-region --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::#UUID#
```

- With aws access key id and aws secret access key options.
```sh
cbmigrate dynamodb --dynamodb-table-name da-test-2 --aws-access-key-id aws-access-key-id --aws-secret-access-key aws-secret-access-key --aws-region aws-region --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name
```

- With hash document key option.
```sh
cbmigrate dynamodb --dynamodb-table-name da-test-2 --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::%firstname%::%lastname% --hash-document-key sha256
```

## Flags

- `--aws-access-key-id string`: AWS Access Key ID.
- `--aws-ca-bundle string`: The CA certificate bundle to use when verifying SSL certificates. Overrides config/env settings.
- `--aws-endpoint-url string`: Override AWS’s default endpoint URL with the given URL.
- `--aws-no-verify-ssl`: By default, the CLI uses SSL when communicating with AWS services. For each SSL connection, the CLI will verify SSL certificates. This option overrides the default behavior of verifying SSL certificates.
- `--aws-profile string`: Use a specific AWS profile from your credential file.
- `--aws-region string`: The region to use. Overrides config/env settings.
- `--aws-secret-access-key string`: AWS Secret Access Key.
- `--buffer-size int`: Buffer size (default 10000).
- `--cb-batch-size int`: Batch size (default 200).
- `--cb-bucket string`: The name of the Couchbase bucket.
- `--cb-cacert string`: Specifies a CA certificate that will be used to verify the identity of the server being connected to. Either this flag or the `--no-ssl-verify` flag must be specified when using an SSL encrypted connection.
- `--cb-client-cert string`: The path to a client certificate used to authenticate when connecting to a cluster. May be supplied with `--client-key` as an alternative to the `--username` and `--password` flags.
- `--cb-client-cert-password string`: The password for the certificate provided to the `--client-cert` flag. When using this flag, the certificate/key pair is expected to be in the PKCS#12 format.
- `--cb-client-key string`: The path to the client private key whose public key is contained in the certificate provided to the `--client-cert` flag. May be supplied with `--client-cert` as an alternative to the `--username` and `--password` flags.
- `--cb-client-key-password string`: The password for the key provided to the `--client-key` flag. When using this flag, the key is expected to be in the PKCS#8 format.
- `--cb-cluster string`: The hostname of a node in the cluster to import data into.
- `--cb-collection string`: The name of the collection where the data needs to be imported. If the collection does not exist, it will be created.
- `--cb-generate-key string`: Specifies a key expression used for generating a key for each document imported. This option allows for the creation of unique document keys in Couchbase by combining static text, field values (denoted by `%fieldname%`), and custom generators (like `#UUID#`) in a format like `"key::%name%::#UUID#"`
- `--cb-no-ssl-verify`: Skips the SSL verification phase. Specifying this flag will allow a connection using SSL encryption but will not verify the identity of the server you connect to. You are vulnerable to a man-in-the-middle attack if you use this flag. Either this flag or the `--cacert` flag must be specified when using an SSL encrypted connection.
- `--cb-password string`: The password for cluster authentication.
- `--cb-scope string`: The name of the scope in which the collection resides. If the scope does not exist, it will be created.
- `--cb-username string`: The username for cluster authentication.
- `--copy-indexes`: Copy indexes for the collection (default true).
- `--debug`: Enable debug output.
- `--dynamodb-table-name string`: The name of the table containing the requested item. You can also provide the Amazon Resource Name (ARN) of the table in this parameter.
- `--dynamodb-limit int`: Specifies the maximum number of items to retrieve per page during a scan operation. Helps control memory usage and API call rates. 
- `--dynamodb-segments int`: Specifies the total number of segments to divide the DynamoDB table into for parallel scanning. Each segment is scanned independently for faster data retrieval. Default is a sequential scan with a single segment (default: 1).
- `-h, --help`: Help for DynamoDB.
- `--hash-document-key string`: Hash the couchbase document key. One of sha256,sha512
- `--keep-primary-key`: Keep the non-composite primary key in the document. By default, if the key is a non-composite primary key, it is deleted from the document unless this flag is set.

## Note
All AWS SDK environment configurations are supported. Click [here](https://docs.aws.amazon.com/sdkref/latest/guide/environment-variables.html) for more info.

For more information about DynamoDB, refer to the following document
- https://docs.aws.amazon.com/dynamodb

For more information about Couchbase, refer to the following document
- https://docs.couchbase.com/home/index.html
