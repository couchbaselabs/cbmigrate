# cbmigrate
A CLI utility to migrate data from other databases into Couchbase.

## Supported Source Databases
Currently, cbmigrate supports migrating data from the following source database:

- **MongoDB**

  For MongoDB migration, cbmigrate provides a specific subcommand. For detailed information on how to use this subcommand, including available options and examples, please refer to the [MongoDB subcommand README](cmd/mongo/README.md).

- **DynamoDB**

  For DynamoDB migration, cbmigrate provides a specific subcommand. For detailed information on how to use this subcommand, including available options and examples, please refer to the [Dynamodb subcommand README](cmd/dynamodb/README.md).

- **HuggingFace**

  For HuggingFace migration, cbmigrate provides a specific subcommand. For detailed information on how to use this subcommand, including available options and examples, please refer to the [HuggingFace subcommand README](cmd/huggingface/README.md).

## Usage
```
cbmigrate [--version] [--help HELP]
cbmigrate [command]
```

### Available Commands
- `dynamodb` - Migrate data from DynamoDB to Couchbase
- `help` - Displays help information about any command
- `mongo` - Migrate data from MongoDB to Couchbase

### Flags
- `-h, --help` - help for `cbmigrate`.
- `-v, --version` - Displays the version of this tool.
