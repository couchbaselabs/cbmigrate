# cbmigrate
A CLI utility to migrate data from other databases into Couchbase.

## Supported Source Databases
Currently, cbmigrate supports migrating data from the following source database:

- **MongoDB**

  For MongoDB migration, cbmigrate provides a specific subcommand. For detailed information on how to use this subcommand, including available options and examples, please refer to the [MongoDB subcommand README](cmd/mongo/README.md).

## Usage
```
cbmigrate [--version] [--help HELP]
cbmigrate [command]
```

### Available Commands
- `help` - Displays help information about any command.
- `mongo` - Specific commands for migrating data from MongoDB.

### Flags
- `-h, --help` - help for `cbmigrate`.
- `-v, --version` - Displays the version of this tool.
