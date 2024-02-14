package repo_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"

	"github.com/couchbaselabs/cbmigrate/internal/common"
	"github.com/couchbaselabs/cbmigrate/internal/couchbase/repo"
)

var _ = Describe("couchbase repo", func() {
	Describe("test group and combine", func() {
		Context("success", func() {
			It("output data should match with the test data with sparse false", func() {
				keys := []common.Key{
					{FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k1", Order: 1},
					{FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k2.n4k1", Order: -1},
					{FieldWithArrayNotation: "k2[].n1k1[].n2k2", Order: 1},
				}
				output, err := repo.GroupAndCombine(keys, false)
				Expect(err).To(BeNil())
				Expect(output).To(Equal("k2[].n1k1[].n2k1.n3k1 ASC INCLUDE MISSING,.n2k1.n3k2.n4k1 DESC INCLUDE MISSING,.n2k2 ASC INCLUDE MISSING"))
			})
			It("output data should match with the test data with sparse true", func() {
				keys := []common.Key{
					{FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k1", Order: 1},
					{FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k2.n4k1", Order: -1},
					{FieldWithArrayNotation: "k2[].n1k1[].n2k2", Order: 1},
				}
				output, err := repo.GroupAndCombine(keys, true)
				Expect(err).To(BeNil())
				Expect(output).To(Equal("k2[].n1k1[].n2k1.n3k1 ASC,.n2k1.n3k2.n4k1 DESC,.n2k2 ASC"))
			})
			It("output data should match with the test data with array in n3", func() {
				keys := []common.Key{
					{FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k1[].n4k1", Order: 1},
					{FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k1[].n4k2.n5k1", Order: -1},
					{FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k1[].n4k2", Order: 1},
				}
				output, err := repo.GroupAndCombine(keys, true)
				Expect(err).To(BeNil())
				Expect(output).To(Equal("k2[].n1k1[].n2k1.n3k1[].n4k1 ASC,.n4k2.n5k1 DESC,.n4k2 ASC"))
			})
		})
		Context("failure", func() {
			It("should return error on multiple array reference", func() {
				keys := []common.Key{
					{FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k1", Order: 1},
					{FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k2.n4k1", Order: -1},
					{FieldWithArrayNotation: "k2[].n1k1[].n2k2[]", Order: 1},
				}
				_, err := repo.GroupAndCombine(keys, false)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("multiple array reference"))
			})
		})
	})
	Describe("generate array expression", func() {
		Context("success", func() {
			It("output data should match with the test data", func() {
				input := "k2[].n1k1[].n2k1.n3k1 ASC INCLUDE MISSING,.n2k1.n3k2.n4k1 DESC INCLUDE MISSING,.n2k2 ASC INCLUDE MISSING"
				output := "DISTINCT ARRAY (DISTINCT ARRAY FLATTEN_KEYS(`l2Item`.`n2k1`.`n3k1` ASC INCLUDE MISSING,`l2Item`.`n2k1`.`n3k2`.`n4k1` DESC INCLUDE MISSING,`l2Item`.`n2k2` ASC INCLUDE MISSING) FOR `l2Item` IN `l1Item`.`n1k1` END) FOR `l1Item` IN `k2` END"
				result := repo.GenerateCouchbaseArrayIndex(input)
				Expect(result).To(Equal(output))
			})
			It("output data should match with the test data", func() {
				input := "k2[].n1k1.n2k1[].n3k1.n4k1[].n5k1 ASC INCLUDE MISSING,.n5k2.n6k1.n7k1 DESC INCLUDE MISSING,.n5k3 ASC INCLUDE MISSING"
				output := "DISTINCT ARRAY (DISTINCT ARRAY (DISTINCT ARRAY FLATTEN_KEYS(`l3Item`.`n5k1` ASC INCLUDE MISSING,`l3Item`.`n5k2`.`n6k1`.`n7k1` DESC INCLUDE MISSING,`l3Item`.`n5k3` ASC INCLUDE MISSING) FOR `l3Item` IN `l2Item`.`n3k1`.`n4k1` END) FOR `l2Item` IN `l1Item`.`n1k1`.`n2k1` END) FOR `l1Item` IN `k2` END"
				result := repo.GenerateCouchbaseArrayIndex(input)
				Expect(result).To(Equal(output))
			})
		})
	})

	Describe("Create index query", func() {
		bucket := "bucket1"
		scope := "scope1"
		collection := "collection1"
		Context("success", func() {
			It("output data should match with the test data", func() {

				index := common.Index{
					Name: "test",
					Keys: []common.Key{
						{Field: "k1.n1k1", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k1", FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k1", Order: 1},
						{Field: "k3", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k2.n4k1", FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k2.n4k1", Order: -1},
						{Field: "k2.n1k1.n2k2", FieldWithArrayNotation: "k2[].n1k1[].n2k2", Order: 1},
						{Field: "k4.n1k2.n2k1", Order: 1},
					},
					Sparse: false,
				}
				arrayExpression := "DISTINCT ARRAY (DISTINCT ARRAY FLATTEN_KEYS(`l2Item`.`n2k1`.`n3k1` ASC INCLUDE MISSING,`l2Item`.`n2k1`.`n3k2`.`n4k1` DESC INCLUDE MISSING,`l2Item`.`n2k2` ASC INCLUDE MISSING) FOR `l2Item` IN `l1Item`.`n1k1` END) FOR `l1Item` IN `k2` END"
				fields := []string{
					"`k1`.`n1k1` ASC INCLUDE MISSING",
					arrayExpression,
					"`k3` ASC INCLUDE MISSING",
					"`k4`.`n1k2`.`n2k1` ASC INCLUDE MISSING",
				}

				Output := fmt.Sprintf(
					"create index %s on `%s`.`%s`.`%s` (%s)",
					index.Name, bucket, scope, collection, strings.Join(fields, ","))
				query, err := repo.CreateIndexQuery(bucket, scope, collection, index)
				Expect(err).To(BeNil())
				Expect(query).To(Equal(Output))
			})
		})
		Context("failure", func() {
			It("output data should match with the test data", func() {
				index := common.Index{
					Name: "test",
					Keys: []common.Key{
						{Field: "k1.n1k1", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k1", FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k1", Order: 1},
						{Field: "k3", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k2.n4k1", FieldWithArrayNotation: "k2[].n1k1[].n2k1.n3k2.n4k1", Order: -1},
						{Field: "k2.n1k1.n2k2", FieldWithArrayNotation: "k2[].n1k1[].n2k2", Order: 1},
						{Field: "k4", FieldWithArrayNotation: "k4[]", Order: 1},
					},
					Sparse: false,
				}
				_, err := repo.CreateIndexQuery(bucket, scope, collection, index)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("multiple array reference"))
			})
		})
	})
})
