package couchbase

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/couchbase/gocb/v2"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/repo"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"strconv"
	"strings"

	"github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	cliErrors "github.com/couchbaselabs/cbmigrate/internal/errors"
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
	db              repo.IRepo
	bucket          string
	scope           string
	collection      string
	batchSize       int
	batchDocs       []gocb.BulkOp
	key             common.ICBDocumentKey
	keepPrimaryKey  bool
	HashDocumentKey string
	processedCount  int
}

type DocKey struct {
	Value string
	Kind  common.DocumentKind // string | field | UUID
}

func NewCouchbase(db repo.IRepo) common.IDestination {
	return &Couchbase{
		db: db,
	}
}

func (c *Couchbase) Init(cbOpts *option.Options, documentKey common.ICBDocumentKey) error {
	c.bucket = cbOpts.Bucket
	c.scope = cbOpts.Scope
	c.collection = cbOpts.Collection
	c.batchSize = cbOpts.BatchSize
	c.key = documentKey
	c.keepPrimaryKey = cbOpts.KeepPrimaryKey
	c.HashDocumentKey = cbOpts.HashDocumentKey
	// The check (only one key is used as a primary key) is needed to for index migration to use meta().ID instead of
	// key while creating the index. Also, that key can be ignored while inserting the doc into couchbase
	var keyParts []common.DocumentKeyPart
	if gk := cbOpts.GeneratedKey; gk != "" {
		splitGK := strings.Split(gk, "::")
		for _, k := range splitGK {
			length := len(k)
			var keyPart common.DocumentKeyPart
			switch {
			case length > 1 && k[0] == '%' && k[length-1] == '%':
				keyPart.Kind = common.DkField
				keyPart.Value = k[1 : length-1]
			case length > 1 && k[0] == '#' && k[length-1] == '#':
				switch k[1 : length-1] {
				case string(common.DkUuid):
					keyPart.Kind = common.DkUuid
					keyPart.Value = k[1 : length-1]
				default:
					return fmt.Errorf("custom generator %s is not supported", k[1:length-1])
				}
			default:
				keyPart.Kind = common.DkString
				keyPart.Value = k
			}
			keyParts = append(keyParts, keyPart)
		}
		c.key.Set(keyParts)
	}
	err := c.db.Init(cbOpts.Cluster, cbOpts)
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
			// only check a collection if scope is found
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
	key := c.key.GetKey()
	kLen := len(key)
	for i, k := range key {
		switch k.Kind {
		case common.DkString:
			id.WriteString(k.Value)
		case common.DkField:
			if val, ok := data[k.Value]; ok {
				id.WriteString(interfaceToString(val))
			}
		case common.DkUuid:
			id.WriteString(getUUID())
		}
		if i < kLen-1 {
			id.WriteString("::")
		}
	}
	if len(key) == 1 && key[0].Kind == common.DkField && c.HashDocumentKey == "" && !c.keepPrimaryKey {
		delete(data, key[0].Value)
	}
	docId := id.String()
	if c.HashDocumentKey != "" {
		var err error
		docId, err = ComputeHash([]byte(id.String()), c.HashDocumentKey)
		if err != nil {
			return err
		}
	}
	c.batchDocs = append(c.batchDocs, &gocb.UpsertOp{
		ID:    docId,
		Value: data,
	})

	// to track the number of documents processed.
	c.processedCount++
	// insert and rest docs when the length of the docs is equal to the batch size
	if len(c.batchDocs)%c.batchSize == 0 {
		err := c.UpsertData()
		if err != nil {
			return err
		}
		zap.S().Infof("%d documents processed", c.processedCount)
		zap.S().Debugf("last processed document %v", id.String())
	}
	return nil
}

func ComputeHash(id []byte, algorithm string) (string, error) {
	switch algorithm {
	case "sha256":
		sha := sha256.Sum256(id)
		return hex.EncodeToString(sha[:]), nil
	case "sha512":
		sha := sha512.Sum512(id)
		return hex.EncodeToString(sha[:]), nil
	}
	return "", fmt.Errorf("hash algorithm: %s not supported", algorithm)
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
	// Be sure to check each operation for errors too.
	for _, op := range c.batchDocs {
		upsertOp := op.(*gocb.UpsertOp)
		if upsertOp.Err != nil {
			zap.S().Errorf("error %#v occured for the document %#v", upsertOp.Err, upsertOp.Value)
		}
	}
	c.batchDocs = nil
	return nil
}

func (c *Couchbase) CreateIndexes(indexes []common.Index) error {
	for _, index := range indexes {
		if index.Error != nil {
			var err cliErrors.NotSupportedError
			if errors.As(index.Error, &err) {
				zap.S().Warnf("error %s occurred while creating index query %s", index.Error.Error(), index.Name)
			} else {
				zap.S().Errorf("error %s occurred while creating index query %s", index.Error.Error(), index.Name)
			}
			continue
		}
		err := c.db.CreateIndex(index.Query)
		if err != nil {
			zap.S().Errorf("error %#v occured while creating index %s", err.Error(), index.Name)
			continue
		}
		zap.S().Debugf("index %s created successfully", index.Name)
	}

	keyspace := fmt.Sprintf("`%s`.`%s`.`%s`", c.bucket, c.scope, c.collection)
	// build differed index
	query := fmt.Sprintf("BUILD INDEX ON %s((SELECT RAW name FROM system:indexes  WHERE "+
		"keyspace_id = '%s' AND scope_id = '%s' AND bucket_id = '%s' AND state = 'deferred' ));",
		keyspace, c.collection, c.scope, c.bucket)
	err := c.db.CreateIndex(query)
	if err != nil {
		zap.S().Errorf("error %#v occured while building indexes", err.Error())
	}
	zap.L().Debug("Indexes deferred are now building in background")
	return nil
}
