package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/dynamodb/option"
	"github.com/couchbaselabs/cbmigrate/internal/dynamodb/repo"
)

type DynamoDB struct {
	collection  string
	db          repo.IRepo
	CopyIndexes bool
	documentKey common.ICBDocumentKey
}

func NewDynamoDB(db repo.IRepo) common.ISource[option.Options] {
	return &DynamoDB{
		db: db,
	}
}

func (d *DynamoDB) Init(opts *option.Options, documentKey common.ICBDocumentKey) error {
	d.documentKey = documentKey
	err := d.db.Init(opts)
	if err != nil {
		return err
	}

	index, err := d.db.GetPrimaryIndex(context.Background())
	if err != nil {
		return err
	}
	var documentKeyParts []common.DocumentKeyPart
	for _, k := range index.Keys {
		documentKeyParts = append(documentKeyParts, common.DocumentKeyPart{
			Value: k,
			Kind:  common.DkField,
		})
	}
	d.documentKey.Set(documentKeyParts)
	// note: the generator key will replace the documentKey settings if present
	return nil
}

func (d *DynamoDB) StreamData(ctx context.Context, mChan chan map[string]interface{}) error {
	defer close(mChan)
	var err error
	paginator := d.db.NewPaginator()
	for paginator.HasMorePages() {
		var output *dynamodb.ScanOutput
		output, err = paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		var records []map[string]interface{}
		err = attributevalue.UnmarshalListOfMaps(output.Items, &records)
		if err != nil {
			return fmt.Errorf("error unmarshalling records %w", err)
		}
		for _, record := range records {
			mChan <- record
		}
	}
	return nil
}

func (d *DynamoDB) GetCouchbaseIndexesQuery(bucket string, scope string, collection string) ([]common.Index, error) {
	indexes, err := d.db.GetIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	isPrimaryIndexPresent := false
	var cbIndexes []common.Index
	for _, index := range indexes {
		query := ""
		switch {
		case len(index.Keys) == 1 && d.documentKey.GetPrimaryKeyOnly() == index.Keys[0]:
			query = fmt.Sprintf(
				"CREATE PRIMARY INDEX `%s` on `%s`.`%s`.`%s` USING GSI WITH {\"defer_build\":true}",
				index.Name, bucket, scope, collection)
			isPrimaryIndexPresent = true
		default:
			query = fmt.Sprintf(
				"CREATE INDEX `%s` on `%s`.`%s`.`%s` (`%s`) USING GSI WITH {\"defer_build\":true}",
				index.Name, bucket, scope, collection, strings.Join(index.Keys, "`,`"))

		}
		cbIndexes = append(cbIndexes, common.Index{
			Name:  index.Name,
			Query: query,
		})
	}
	if !isPrimaryIndexPresent {
		uuid, _ := common.GenerateShortUUIDHex()
		key := "primary-" + uuid
		index := common.Index{
			Name: key,
			Query: fmt.Sprintf(
				"CREATE PRIMARY INDEX `%s` on `%s`.`%s`.`%s` USING GSI WITH {\"defer_build\":true}",
				key, bucket, scope, collection),
		}
		cbIndexes = append(cbIndexes, index)
	}
	return cbIndexes, nil
}
