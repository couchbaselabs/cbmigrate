package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"go.uber.org/zap"
	"strings"
	"sync"

	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/dynamodb/option"
	"github.com/couchbaselabs/cbmigrate/internal/dynamodb/repo"
)

type DynamoDB struct {
	collection  string
	db          repo.IRepo
	CopyIndexes bool
	documentKey common.ICBDocumentKey
	segments    int
	limit       int
}

func NewDynamoDB(db repo.IRepo) common.ISource[option.Options] {
	return &DynamoDB{
		db: db,
	}
}

func (d *DynamoDB) Init(opts *option.Options, documentKey common.ICBDocumentKey) error {
	d.documentKey = documentKey
	d.segments = opts.Segments
	d.limit = opts.Limit
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

	errChan := make(chan error, d.segments)
	var wg sync.WaitGroup
	dCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	for segment := 0; segment < d.segments; segment++ {
		wg.Add(1)
		go func(segment int) {
			defer wg.Done()
			err := d.parallelScanSegment(dCtx, segment, mChan)
			if err != nil {
				sync.OnceFunc(func() {
					cancel()
				})
				errChan <- err
			}
		}(segment)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}
	if errs != nil {
		zap.L().Debug(errors.Join(errs...).Error())
		return errs[0]
	}
	return nil
}

func (d *DynamoDB) parallelScanSegment(ctx context.Context, segment int, mChan chan map[string]interface{}) error {
	paginator := d.db.NewPaginator(int32(segment), int32(d.segments), int32(d.limit))
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		var records []map[string]interface{}
		err = attributevalue.UnmarshalListOfMaps(output.Items, &records)
		if err != nil {
			return fmt.Errorf("error unmarshalling records in segment %d: %w", segment, err)
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
	ncpk := d.documentKey.GetNonCompoundPrimaryKeyOnly()
	for _, index := range indexes {

		query := ""
		switch {
		case len(index.Keys) == 1 && d.documentKey.GetNonCompoundPrimaryKeyOnly() == index.Keys[0]:
			query = fmt.Sprintf(
				"CREATE PRIMARY INDEX `%s` on `%s`.`%s`.`%s` USING GSI WITH {\"defer_build\":true}",
				index.Name, bucket, scope, collection)
			isPrimaryIndexPresent = true
		default:
			keys := make([]string, len(index.Keys))
			for i, k := range index.Keys {
				// this condition is added to point the non-compound primary key to meta().id as the it will be removed
				// in the document to avoid redundant information
				if k == ncpk {
					keys[i] = common.MetaDataID
				} else {
					keys[i] = k
				}
			}
			query = fmt.Sprintf(
				"CREATE INDEX `%s` on `%s`.`%s`.`%s` (`%s`) USING GSI WITH {\"defer_build\":true}",
				index.Name, bucket, scope, collection, strings.Join(keys, "`,`"))

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
