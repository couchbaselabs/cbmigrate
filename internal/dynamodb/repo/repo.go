package repo

//go:generate mockgen -source=repo.go -destination=../../../testhelper/mock/dynamodb_repo.go -package=mock -mock_names=IRepo=MockDynamoDbIRepo,IPaginator=MockDynamoDbIPaginator IRepo IPaginator

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamoDB "github.com/couchbaselabs/cbmigrate/internal/db/dynamodb"
	"github.com/couchbaselabs/cbmigrate/internal/dynamodb/option"
	"strings"
)

type IRepo interface {
	Init(opts *option.Options) error
	NewPaginator(segment int32, totalSegments int32, limit int32) IPaginator
	GetIndexes(ctx context.Context) ([]Index, error)
	GetPrimaryIndex(ctx context.Context) (Index, error)
}

type Index struct {
	Name string
	Keys []string
}

type IPaginator interface {
	HasMorePages() bool
	NextPage(ctx context.Context, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

type Repo struct {
	TableName string
	svc       *dynamoDB.DB
}

func NewRepo() IRepo {
	return &Repo{
		svc: new(dynamoDB.DB),
	}
}

func (r *Repo) Init(opts *option.Options) error {
	r.TableName = opts.TableName
	return r.svc.Init(opts)
}

func (r *Repo) NewPaginator(segment int32, totalSegments int32, limit int32) IPaginator {
	si := &dynamodb.ScanInput{
		TableName:     aws.String(r.TableName),
		Segment:       &segment,
		TotalSegments: &totalSegments,
	}
	if limit > 0 {
		si.Limit = &limit
	}
	return dynamodb.NewScanPaginator(r.svc, si)
}

func (r *Repo) GetIndexes(ctx context.Context) ([]Index, error) {
	output, err := r.svc.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(r.TableName)})
	if err != nil {
		return nil, err
	}
	var indexes []Index
	indexes = append(indexes, getIndexFromSchema(output.Table.KeySchema))
	for _, index := range output.Table.LocalSecondaryIndexes {
		indexes = append(indexes, getIndexFromSchema(index.KeySchema))
	}
	for _, index := range output.Table.GlobalSecondaryIndexes {
		indexes = append(indexes, getIndexFromSchema(index.KeySchema))
	}
	return indexes, nil
}

func (r *Repo) GetPrimaryIndex(ctx context.Context) (Index, error) {
	output, err := r.svc.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(r.TableName)})
	if err != nil {
		return Index{}, err
	}
	return getIndexFromSchema(output.Table.KeySchema), nil
}

func getIndexFromSchema(kse []types.KeySchemaElement) Index {
	var name strings.Builder
	var keys []string
	for i, ks := range kse {
		if i != 0 {
			name.WriteString("-")
		}
		name.WriteString(*ks.AttributeName)
		keys = append(keys, *ks.AttributeName)
	}
	return Index{
		Name: name.String(),
		Keys: keys,
	}
}
