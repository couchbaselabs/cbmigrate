
## Overview

`cbmigrate` is a CLI tool designed for migrating data from MongoDB to Couchbase. It supports a variety of features to make the migration process seamless and customizable, including SSL encryption options, custom key generation for documents, and the ability to specify the target bucket, scope, and collection in Couchbase.

## Features

- Direct migration from MongoDB to Couchbase.
- SSL encryption support with optional verification.
- Customizable document key generation.
- Option to copy MongoDB indexes with considerations for specific types.
- Verbose output for detailed operation logs.


## Usage:
```
cbmigrate mongo --mongodb-uri MONGODB_URI --mongodb-collection MONGODB_COLLECTION --mongodb-database MONGODB_DATABASE --cb-cluster CB_CLUSTER (--cb-username CB_USERNAME --cb-password CB_PASSWORD | --cb-client-cert CB_CLIENT_CERT [--cb-client-cert-password CB_CLIENT_CERT_PASSWORD] [--cb-client-key CB_CLIENT_KEY] [--cb-client-key-password CB_CLIENT_KEY_PASSWORD]) [--cb-generate-key CB_GENERATE_KEY] [--cb-cacert CB_CACERT] [--cb-no-ssl-verify CB_NO_SSL_VERIFY] [--cb-bucket CB_BUCKET] [--cb-scope CB_SCOPE] [--cb-collection CB_COLLECTION] [--verbose] [--copy-indexes] [--help HELP]
```

## Aliases:
- mongo, m

## Examples:
- Importing data from MongoDB to Couchbase:
  ```
  cbmigrate mongo --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name
  ```
- Using dot notation for nested fields:
  ```
  cbmigrate mongo --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::%name.first_name%::%name.last_name%
  ```
- Generating keys with UUID:
  ```
  cbmigrate mongo --mongodb-uri uri --mongodb-database db-name --mongodb-collection collection-name --cb-cluster url --cb-username username --cb-password password --cb-bucket bucket-name --cb-scope scope-name --cb-collection collection-name --cb-generate-key key::#UUID#
  ```

## Flags:
- `--cb-bucket string`: The name of the Couchbase bucket.
- `--cb-cacert string`: Specifies a CA certificate that will be used to verify the identity of the server being connecting to. Either this flag or the --cb-no-ssl-verify flag must be specified when using an SSL encrypted connection.
- `--cb-client-cert string`: The path to a client certificate used to authenticate when connecting to a cluster. Maybe supplied with --client-key as an alternative to the --cb-username and --cb-password flags.
- `--cb-client-cert-password string`: The password for the certificate provided to the --client-cert flag, when using this flag, the certificate/key pair is expected to be in the PKCS#12 format.
- `--cb-client-key string`: The path to the client private key whose public key is contained in the certificate provided to the --client-cert flag. May be supplied with --client-cert as an alternative to the --cb-username and --cb-password flags.
- `--cb-client-key-password string`: The password for the key provided to the --client-key flag, when using this flag, the key is expected to be in the PKCS#8 format.
- `--cb-cluster string`: The hostname of a node in the cluster to import data into.
- `--cb-collection string`: The name of the collection where the data needs to be imported. If the collection does not exist, it will be created.
- `--cb-generate-key string`: Specifies a key expression used for generating a key for each document imported. This option allows for the creation of unique document keys in Couchbase by combining static text, field values (denoted by %fieldname%), and custom generators (like #UUID#) in a format like "key::%name%::#UUID#" (default "%_id%")
- `--cb-no-ssl-verify`: Skips the SSL verification phase. Specifying this flag will allow a connection using SSL encryption, but will not verify the identity of the server you connect to. You are vulnerable to a man-in-the-middle attack if you use this flag. Either this flag or the --cacert flag must be specified when using an SSL encrypted connection.
- `--cb-password string`: The password for cluster authentication.
- `--cb-scope string`: The name of the scope in which the collection resides. If the scope does not exist, it will be created.
- `--cb-username string`: The username for cluster authentication.
- `--copy-indexes`:Copy indexes for the collection (default true).
- `--help`: help for mongo
- `--mongodb-collection string`: MongoDB collection to use.
- `--mongodb-database string`: MongoDB database to use.
- `--mongodb-uri string`: MongoDB URI connection string.
- `--verbose`: Enable verbose output.



## Index Translation: MongoDB to Couchbase

### Example Document
```json
{
  "user_id": "45678",
  "name": "Elena Rodriguez",
  "birthdate": "1992-04-16T00:00:00Z",
  "professional_skills": ["JavaScript", "Python", "React", "Node.js", "SQL"],
  "projects_experience": [
    {
      "project_name": "Web Application for E-commerce",
      "role": "Lead Developer",
      "duration": {
        "start": "2021-05-01T00:00:00Z",
        "end": "2022-03-31T00:00:00Z"
      },
      "technologies_used": ["React", "Node.js", "MongoDB", "Express"],
      "description": "Developed a full-stack web application to support e-commerce operations, integrating payment gateways and managing user data securely."
    },
    {
      "project_name": "Data Analysis Platform",
      "role": "Data Engineer",
      "duration": {
        "start": "2022-04-01T00:00:00Z"
      },
      "technologies_used": ["Python", "Pandas", "SQL", "Docker"],
      "description": "Designed and implemented a platform for analyzing large datasets, optimizing data processing and visualization for business insights."
    }
  ],
  "educational_background": {
    "degree": "Master of Science in Computer Science",
    "institution": "Tech University",
    "graduation_date": "2017-05-20T00:00:00Z",
    "thesis_title": "Optimizing Database Performance with Machine Learning",
    "advisor": "Dr. Sarah Connors"
  },
  "contact_details": {
    "email": "elena.rodriguez@example.com",
    "phone": "+1234567890",
    "address": {
      "street": "123 Tech Drive",
      "city": "Innovate City",
      "state": "Techland",
      "zip": "98765"
    }
  }
}
```

### Scenario 1  Creating a compound index
#### MongoDB Index Creation
In MongoDB, you might create a compound index on a collection myColl that involves top-level fields, fields within a sub-document and nested fields within an array of sub-document, as shown below:
```mongodb
db.myColl.createIndex({ "user_id": 1, "projects_experience.duration.start": 1, "contact_details.email": 1, "projects_experience.role": 1 })
```

#### Couchbase Index Translation
In Couchbase, to accommodate the same indexing structure, the syntax needs to be adapted to handle arrays explicitly. The equivalent Couchbase command uses the **ARRAY** keyword and the **FOR** loop construct to iterate over the array elements, applying the indexing to each nested field within the array. Here's how the Couchbase command might look:
```couchbase
CREATE INDEX example_index ON myColl(user_id ASC, ALL ARRAY FLATTEN_KEYS(l1Item.duration.start, l1Item.role) FOR l1Item IN projects_experience END, contact_details.email ASC)
```
In this Couchbase index creation command:

- **user_id** ASC and **contact_details.email** ASC directly translate the top-level fields **user_id** and **contact_details.email** from the MongoDB index.
- **ALL ARRAY FLATTEN_KEYS(l1Item.duration.start, l1Item.role) FOR l1Item IN projects_experience END** is used to index the nested fields within the array **projects_experience**, specifically the paths **duration.start** and **role** for each item in **projects_experience**. This ensures that each element within the array **projects_experience** is indexed according to the nested fields specified.


### Scenario 2 Creating an index with partial filter expression
#### MongoDB Index Creation
In MongoDB, you might create an index on a collection myColl that involves top-level fields, fields within a sub-document and nested fields within an array of sub-document, as shown below:
```mongodb
db.myColl.createIndex({ "user_id": 1},{"partialFilterExpression": {"user_id":100,"projects_experience.duration.start" : {"$gte": new Date("2021-10-12")}}, "projects_experience.duration.end" : {"$lte": new Date("2022-10-12")}},"contact_details.email" : {"$type": "string"}})
```

#### Couchbase Index Translation
In Couchbase, to accommodate the same indexing structure, the syntax needs to be adapted to handle arrays explicitly. The equivalent Couchbase command uses the **ARRAY** keyword and the **FOR** loop construct to iterate over the array elements, applying the indexing to each nested field within the array. Here's how the Couchbase command might look:
```couchbase
CREATE INDEX example_index ON myColl(user_id INCLUDE MISSING ASC) WHERE (`user_id` = 100 AND ANY `l1Item` IN `projects_experience` SATISFIES (`l2Item`.`duration`.`start` >= 2021-10-12T00:00:00Z ) END AND ANY `l1Item` IN `projects_experience` SATISFIES (`l2Item`.`duration`.`end` <= 2022-10-12T00:00:00Z ) END AND type(`contact_details`.`email`) = "string")
```
In this Couchbase index creation command:

- **user_id** and **contact_details.email** directly translate the top-level fields **user_id** and **contact_details.email** from the MongoDB index.
- **ANY `l1Item` IN `projects_experience` SATISFIES (`l2Item`.`duration`.`start` >= 2021-10-12T00:00:00Z ) END** is used to represent the nested field within the array **projects_experience** in partial expression.
- **ANY `l1Item` IN `projects_experience` SATISFIES (`l2Item`.`duration`.`end` <= 2022-10-12T00:00:00Z ) END** is used to represent the nested field within the array **projects_experience** in partial expression.


## Limitations

- Date and decimal types in MongoDB are converted to strings in Couchbase. Date string is in RFC3339 format.
- While migrating the indexes currently text, wildcard and indexes with collation are currently not supported.
- Compound index translations involving arrays and objects require specific syntax adaptations.

For more information about Mongodb indexes, refer to the following docs
- https://www.mongodb.com/docs/manual/indexes
- https://www.mongodb.com/docs/manual/core/indexes/index-types
- https://www.mongodb.com/docs/manual/applications/indexes

For more information about Couchbase indexes, refer to the following docs
- https://docs.couchbase.com/server/current/learn/services-and-indexes/indexes/indexes.html
- https://docs.couchbase.com/server/current/learn/services-and-indexes/indexes/global-secondary-indexes.html
- https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/createindex.html