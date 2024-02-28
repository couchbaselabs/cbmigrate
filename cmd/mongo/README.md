
## Overview

`cbmigrate` is a CLI tool designed for migrating data from MongoDB to Couchbase. It supports a variety of features to make the migration process seamless and customizable, including SSL encryption options, custom key generation for documents, and the ability to specify the target bucket, scope, and collection in Couchbase.

## Features

- Direct migration from MongoDB to Couchbase.
- SSL encryption support with optional verification.
- Customizable document key generation.
- Option to copy MongoDB indexes with considerations for specific types.
- Verbose output for detailed operation logs.

## Installation

*Installation instructions specific to `cbmigrate` should be provided here, including any prerequisites, dependencies, and a step-by-step guide.*

## Usage

### Command Syntax

```
cbmigrate mongo --mongodb-uri MONGODB_URI --mongodb-collection MONGODB_COLLECTION --mongodb-database MONGODB_DATABASE --cb-cluster CB_CLUSTER [Authentication Options] [Connection Options] [Data Mapping Options] [--verbose] [--copy-indexes]
```

### Examples

1. **Basic Import Example**

```
cbmigrate mongo --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name
```

2. **Custom Key Generation Example**

```
cbmigrate mongo --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::%name.first_name%::%name.last_name%
```

3. **UUID Key Generation Example**

```
cbmigrate mongo --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::#UUID#
```

### Flags

*Detailed explanations of all available flags should be provided here.*

## Limitations

- Date and decimal types in MongoDB are converted to strings in Couchbase.
- While migrating the indexes currently text, wildcard and indexes with collation are currently not supported.
- Compound index translations involving arrays and objects require specific syntax adaptations.

## Index Translation Example: MongoDB to Couchbase

### MongoDB Index Creation
In MongoDB, you might create a compound index on a collection myColl that involves both top-level fields and nested fields within an array, as shown below:
```mongodb
db.myColl.createIndex({ k1: 1, "k2.n1k1.n2k1": 1, k3: 1, "k2.n1k1.n2k2": 1 })
```
This index includes top-level fields (k1, k3) and nested fields within an array (k2.n1k1.n2k1, k2.n1k1.n2k2),here k2 is Array of objects.

### Couchbase Index Translation
In Couchbase, to accommodate the same indexing structure, the syntax needs to be adapted to handle arrays explicitly. The equivalent Couchbase command utilizes the **ARRAY** keyword and the **FOR** loop construct to iterate over the array elements, applying the indexing to each nested field within the array. Here's how the Couchbase command might look:
```couchbase
CREATE INDEX example_index ON myColl(k1 ASC, ALL ARRAY FLATTEN_KEYS(l1Item.n1k1.n2k1, l1Item.n1k1.n2k2) FOR l1Item IN k2 END, k3 ASC)
```
In this Couchbase index creation command:

- **k1** ASC and **k3** ASC directly translate the top-level fields **k1** and **k3** from the MongoDB index.
- **ALL ARRAY FLATTEN_KEYS(l1Item.n1k1.n2k1, l1Item.n1k1.n2k2) FOR l1Item IN k2 END** is used to index the nested fields within the array **k2**, specifically the paths **n1k1.n2k1** and **n1k1.n2k2** for each item in **k2**. This ensures that each element within the array **k2** is indexed according to the nested fields specified.
