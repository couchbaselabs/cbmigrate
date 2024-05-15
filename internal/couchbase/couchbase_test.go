package couchbase_test

import (
	"github.com/couchbase/gocb/v2"
	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase"
	cOpts "github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	mock_test "github.com/couchbaselabs/cbmigrate/testhelper/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	"reflect"
)

var scopeSpec1 = gocb.ScopeSpec{
	Name: "test_scope",
	Collections: []gocb.CollectionSpec{
		{
			Name:      "test_col",
			ScopeName: "test_scope",
		},
	},
}
var scopeSpec2 = gocb.ScopeSpec{
	Name: "scope2",
	Collections: []gocb.CollectionSpec{
		{
			Name:      "col2",
			ScopeName: "scope2",
		},
	},
}
var scopeSpec3 = gocb.ScopeSpec{
	Name: "scope3",
	Collections: []gocb.CollectionSpec{
		{
			Name:      "col3",
			ScopeName: "scope3",
		},
	},
}

var _ = Describe("couchbase service", func() {
	Describe("test couchbase connection initialization", func() {
		var (
			ctrl             *gomock.Controller
			db               *mock_test.MockCouchbaseIRepo
			couchbaseService common.IDestination
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			db = mock_test.NewMockCouchbaseIRepo(ctrl)
			couchbaseService = couchbase.NewCouchbase(db)
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		Context("init connection success", func() {
			It("scope and collection exists", func() {
				opts := &cOpts.Options{
					Cluster:      "cluster-url",
					NameSpace:    &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col"},
					BatchSize:    100,
					GeneratedKey: "%_id%",
				}
				db.EXPECT().Init(opts.Cluster, opts).Return(nil)
				db.EXPECT().GetAllScopes().DoAndReturn(func() ([]gocb.ScopeSpec, error) {
					return []gocb.ScopeSpec{
						scopeSpec1,
						scopeSpec2,
						scopeSpec3,
					}, nil
				})
				docKey := common.NewCBDocumentKey()
				docKey.Set([]common.DocumentKeyPart{{Kind: common.DkField, Value: "id"}})
				err := couchbaseService.Init(opts, docKey)
				Expect(docKey.GetKey()).To(Equal([]common.DocumentKeyPart{{Kind: common.DkField, Value: "_id"}}))
				Expect(err).To(BeNil())

			})
			It("scope and collection exists and without generated key", func() {
				opts := &cOpts.Options{
					Cluster:   "cluster-url",
					NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col"},
					BatchSize: 100,
				}
				db.EXPECT().Init(opts.Cluster, opts).Return(nil)
				db.EXPECT().GetAllScopes().DoAndReturn(func() ([]gocb.ScopeSpec, error) {
					return []gocb.ScopeSpec{
						scopeSpec1,
						scopeSpec2,
						scopeSpec3,
					}, nil
				})
				docKey := common.NewCBDocumentKey()
				docKey.Set([]common.DocumentKeyPart{{Kind: common.DkField, Value: "id"}})
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(docKey.GetKey()).To(Equal([]common.DocumentKeyPart{{Kind: common.DkField, Value: "id"}}))
				Expect(err).To(BeNil())

			})
			It("create scope and collection", func() {
				opts := &cOpts.Options{
					Cluster:   "cluster-url",
					NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col"},
					BatchSize: 100,
				}
				db.EXPECT().Init(opts.Cluster, opts).Return(nil)
				db.EXPECT().GetAllScopes().DoAndReturn(func() ([]gocb.ScopeSpec, error) {
					return []gocb.ScopeSpec{
						scopeSpec2,
						scopeSpec3,
					}, nil
				})
				db.EXPECT().CreateScope(opts.Scope).Return(nil)
				db.EXPECT().CreateCollection(opts.Scope, opts.Collection).Return(nil)
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).To(BeNil())
			})
			It("create collection", func() {
				opts := &cOpts.Options{
					Cluster:   "cluster-url",
					NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col2"},
					BatchSize: 100,
				}

				db.EXPECT().Init(opts.Cluster, opts).Return(nil)
				db.EXPECT().GetAllScopes().DoAndReturn(func() ([]gocb.ScopeSpec, error) {
					return []gocb.ScopeSpec{
						scopeSpec1,
						scopeSpec2,
						scopeSpec3,
					}, nil
				})
				db.EXPECT().CreateCollection(opts.Scope, opts.Collection).Return(nil)
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).To(BeNil())
			})
		})
		Context("init connection failure", func() {
			It("get all scopes", func() {
				opts := &cOpts.Options{
					Cluster:   "cluster-url",
					NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col"},
					BatchSize: 100,
				}
				getAllScopeError := errors.New("error in getting all scopes")
				db.EXPECT().Init(opts.Cluster, opts).Return(nil)
				db.EXPECT().GetAllScopes().Return(nil, getAllScopeError)
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(getAllScopeError))
			})
			It("create scopes", func() {
				opts := &cOpts.Options{
					Cluster:   "cluster-url",
					NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col"},
					BatchSize: 100,
				}
				createScopeError := errors.New("error in creating scope")
				db.EXPECT().Init(opts.Cluster, opts).Return(nil)
				db.EXPECT().GetAllScopes().DoAndReturn(func() ([]gocb.ScopeSpec, error) {
					return []gocb.ScopeSpec{
						scopeSpec2,
						scopeSpec3,
					}, nil
				})
				db.EXPECT().CreateScope(opts.Scope).Return(createScopeError)
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(createScopeError))
			})
			It("create collection with existing scope", func() {
				opts := &cOpts.Options{
					Cluster:   "cluster-url",
					NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col1"},
					BatchSize: 100,
				}
				createCollectionError := errors.New("error in creating collection")
				db.EXPECT().Init(opts.Cluster, opts).Return(nil)
				db.EXPECT().GetAllScopes().DoAndReturn(func() ([]gocb.ScopeSpec, error) {
					return []gocb.ScopeSpec{
						scopeSpec1,
						scopeSpec2,
						scopeSpec3,
					}, nil
				})
				db.EXPECT().CreateCollection(opts.Scope, opts.Collection).Return(createCollectionError)
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(createCollectionError))
			})
			It("create collection with non existing scope", func() {
				opts := &cOpts.Options{
					Cluster:   "cluster-url",
					NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope1", Collection: "test_col1"},
					BatchSize: 100,
				}
				createCollectionError := errors.New("error in creating collection")
				db.EXPECT().Init(opts.Cluster, opts).Return(nil)
				db.EXPECT().GetAllScopes().DoAndReturn(func() ([]gocb.ScopeSpec, error) {
					return []gocb.ScopeSpec{
						scopeSpec1,
						scopeSpec2,
						scopeSpec3,
					}, nil
				})
				db.EXPECT().CreateScope(opts.Scope).Return(nil)
				db.EXPECT().CreateCollection(opts.Scope, opts.Collection).Return(createCollectionError)
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(createCollectionError))
			})
			It("Error in initializing db connection", func() {
				opts := &cOpts.Options{
					Cluster:   "cluster-url",
					NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col1"},
					BatchSize: 100,
				}
				dbConInitError := errors.New("error in initializing db connection")
				db.EXPECT().Init(opts.Cluster, opts).Return(dbConInitError)
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(dbConInitError))
			})
		})
	})

	Describe("test couchbase data processing", func() {
		var (
			ctrl             *gomock.Controller
			db               *mock_test.MockCouchbaseIRepo
			couchbaseService common.IDestination
			docs             []map[string]interface{}
		)
		opts := &cOpts.Options{
			Cluster:   "cluster-url",
			NameSpace: &cOpts.NameSpace{Bucket: "test_bucket", Scope: "test_scope", Collection: "test_col"},
			BatchSize: 100,
		}
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			db = mock_test.NewMockCouchbaseIRepo(ctrl)
			couchbaseService = couchbase.NewCouchbase(db)
			db.EXPECT().Init(opts.Cluster, opts).Return(nil)
			db.EXPECT().GetAllScopes().DoAndReturn(func() ([]gocb.ScopeSpec, error) {
				return []gocb.ScopeSpec{
					scopeSpec1,
					scopeSpec2,
					scopeSpec3,
				}, nil
			})
			docs = make([]map[string]interface{}, 500)
			for i := 0; i < 500; i++ {
				data := map[string]interface{}{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
					"k4": "v4",
					"k5": "v5",
				}
				data["id"] = i + 1
				docs[i] = data
			}
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		Context("data processing success", func() {
			It("upsert data when batch size is 100", func() {
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).To(BeNil())
				i := 0
				db.EXPECT().UpsertData(opts.Scope, opts.Collection, gomock.Any()).Times(5).DoAndReturn(func(scope, collection string, uDocs []gocb.BulkOp) error {
					for _, d := range uDocs {
						if !reflect.DeepEqual(d.(*gocb.UpsertOp).Value, docs[i]) {
							return errors.New("doc not equal")
						}
						i++
					}
					return nil
				})
				for _, doc := range docs {
					err = couchbaseService.ProcessData(doc)
					Expect(err).To(BeNil())
				}
				err = couchbaseService.Complete()
				Expect(err).To(BeNil())
			})
			It("upsert data when batch size is 100 and call complete option", func() {
				docsLen := len(docs)
				for i := 0; i < 50; i++ {
					docs = append(docs, map[string]interface{}{
						"id ": docsLen + i + 1,
						"k1":  "v1",
						"k2":  "v2",
						"k3":  "v3",
						"k4":  "v4",
						"k5":  "v5",
					})
				}
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).To(BeNil())
				i := 0
				db.EXPECT().UpsertData(opts.Scope, opts.Collection, gomock.Any()).Times(6).DoAndReturn(func(scope, collection string, uDocs []gocb.BulkOp) error {
					for _, d := range uDocs {
						if !reflect.DeepEqual(d.(*gocb.UpsertOp).Value, docs[i]) {
							return errors.New("doc not equal")
						}
						i++
					}
					return nil
				})
				for _, doc := range docs {
					err = couchbaseService.ProcessData(doc)
					Expect(err).To(BeNil())
				}
				err = couchbaseService.Complete()
				Expect(err).To(BeNil())
			})
		})
		Context("data processing failure", func() {
			It("upsert data failure in process data function and complete function", func() {
				docsLen := len(docs)
				for i := 0; i < 50; i++ {
					docs = append(docs, map[string]interface{}{
						"id ": docsLen + i + 1,
						"k1":  "v1",
						"k2":  "v2",
						"k3":  "v3",
						"k4":  "v4",
						"k5":  "v5",
					})
				}
				err := couchbaseService.Init(opts, common.NewCBDocumentKey())
				Expect(err).To(BeNil())
				processDataError := errors.New("error in processing the data")
				db.EXPECT().UpsertData(opts.Scope, opts.Collection, gomock.Any()).Times(6).Return(processDataError)
				err = couchbaseService.ProcessData(docs[0])
				for i, doc := range docs {
					err = couchbaseService.ProcessData(doc)
					if i+1%opts.BatchSize == 0 {
						Expect(err).NotTo(BeNil())
						Expect(err).To(Equal(processDataError))
					}
				}
				err = couchbaseService.Complete()
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(processDataError))
			})
		})
	})
})
