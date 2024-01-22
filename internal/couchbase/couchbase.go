package couchbase

import (
	"fmt"
	"github.com/couchbase/gocb/v2"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/repo"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"strconv"
	"strings"

	"github.com/couchbaselabs/cbmigrate/internal/option"
)

func interfaceToString(value interface{}) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case primitive.ObjectID:
		return v.Hex()
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getUUID() string {
	var id uuid.UUID
	var err error
	retry := 5
	for i := 0; i <= retry; i++ {
		id, err = uuid.NewRandom()
		if err != nil {
			continue
		}
		break
	}
	return id.String()

}

type Couchbase struct {
	db             repo.IRepo
	bucket         string
	scope          string
	collection     string
	batchSize      int
	batchDocs      []gocb.BulkOp
	key            []docKey
	processedCount int
}

type docKey struct {
	value string
	kind  string // string  | field | UUID
}

func NewCouchbase(db repo.IRepo) common.IDestination {
	return &Couchbase{
		db: db,
	}
}

func (c *Couchbase) Init(opts *option.Options) error {
	cbOpts := opts.CBOpts
	c.bucket = cbOpts.Bucket
	c.scope = cbOpts.Scope
	c.collection = cbOpts.Collection
	c.batchSize = cbOpts.BatchSize
	if gk := cbOpts.GeneratedKey; gk != "" {
		splitGK := strings.Split(gk, "::")

		for _, k := range splitGK {
			length := len(k)
			switch {
			case length > 1 && k[0] == '%' && k[length-1] == '%':
				c.key = append(c.key, docKey{kind: "field", value: k[1 : length-1]})
			case k == "UUID":
				c.key = append(c.key, docKey{kind: "UUID", value: k})
			default:
				c.key = append(c.key, docKey{kind: "string", value: k})
			}
		}
	}
	err := c.db.Init(cbOpts.Cluster, opts.CBOpts)
	if err != nil {
		return err
	}
	return c.createScopeAndCollectionIFNotExits()
}

func (c *Couchbase) createScopeAndCollectionIFNotExits() error {
	foundScope := false
	foundCollection := false

	scopes, err := c.db.GetAllScopes()
	if err != nil {
		return err
	}
l1:
	for _, scope := range scopes {
		if scope.Name == c.scope {
			foundScope = true
		} else {
			// only check collection if scope is found
			continue
		}

		for _, col := range scope.Collections {
			if col.Name == c.collection {
				foundCollection = true
				break l1
			}
		}
	}
	if foundScope != true {
		err = c.db.CreateScope(c.scope)
		if err != nil {
			return err
		}
	}
	if foundCollection != true {
		err = c.db.CreateCollection(c.scope, c.collection)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Couchbase) ProcessData(data map[string]interface{}) error {
	var id strings.Builder
	for _, k := range c.key {
		switch k.kind {
		case "string":
			id.WriteString(k.value)
		case "field":
			if val, ok := data[k.value]; ok {
				id.WriteString(interfaceToString(val))
			}
		case "UUID":
			id.WriteString(getUUID())
		}
	}
	c.batchDocs = append(c.batchDocs, &gocb.UpsertOp{
		ID:    id.String(),
		Value: data,
	})

	// to track the number of documents processed.
	c.processedCount++
	// insert and rest docs when length of the docs is equal to the batch size
	if len(c.batchDocs)%c.batchSize == 0 {
		err := c.UpsertData()
		if err != nil {
			return err
		}
		zap.S().Infof("%d documents processed", c.processedCount)
		zap.S().Infof("last processed document %v", id.String())
	}
	return nil
}

func (c *Couchbase) Complete() (err error) {
	if len(c.batchDocs) == 0 {
		return nil
	}
	return c.UpsertData()
}

func (c *Couchbase) UpsertData() error {
	err := c.db.UpsertData(c.scope, c.collection, c.batchDocs)
	if err != nil {
		return err
	}
	// Be sure to check each individual operation for errors too.
	for _, op := range c.batchDocs {
		upsertOp := op.(*gocb.UpsertOp)
		if upsertOp.Err != nil {
			zap.S().Errorf("error %#v occured for the document %#v", upsertOp.Err, upsertOp.Value)
		}
	}
	c.batchDocs = nil
	return nil
}
