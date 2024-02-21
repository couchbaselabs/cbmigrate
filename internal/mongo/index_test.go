package mongo_test

import (
	"fmt"
	"github.com/couchbaselabs/cbmigrate/internal/mongo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("mongo to couchbase index ", func() {
	Describe("test group and combine", func() {
		Context("success", func() {
			It("output data should match with the test data with sparse false", func() {
				keys := []mongo.Key{
					{Field: "k2[].n1k1[].n2k1.n3k1", Order: 1},
					{Field: "k2[].n1k1[].n2k1.n3k2.n4k1", Order: -1},
					{Field: "k2[].n1k1[].n2k2", Order: 1},
				}
				output, err := mongo.GroupAndCombine(keys, true)
				Expect(err).To(BeNil())
				Expect(output).To(Equal("k2[].n1k1[].n2k1.n3k1 ASC INCLUDE MISSING,.n2k1.n3k2.n4k1 DESC,.n2k2 ASC"))
			})
			It("output data should match with the test data with sparse true", func() {
				keys := []mongo.Key{
					{Field: "k2[].n1k1[].n2k1.n3k1", Order: 1},
					{Field: "k2[].n1k1[].n2k1.n3k2.n4k1", Order: -1},
					{Field: "k2[].n1k1[].n2k2", Order: 1},
				}
				output, err := mongo.GroupAndCombine(keys, false)
				Expect(err).To(BeNil())
				Expect(output).To(Equal("k2[].n1k1[].n2k1.n3k1 ASC,.n2k1.n3k2.n4k1 DESC,.n2k2 ASC"))
			})
			It("output data should match with the test data with array in n3", func() {
				keys := []mongo.Key{
					{Field: "k2[].n1k1[].n2k1.n3k1[].n4k1", Order: 1},
					{Field: "k2[].n1k1[].n2k1.n3k1[].n4k2.n5k1", Order: -1},
					{Field: "k2[].n1k1[].n2k1.n3k1[].n4k2", Order: 1},
				}
				output, err := mongo.GroupAndCombine(keys, false)
				Expect(err).To(BeNil())
				Expect(output).To(Equal("k2[].n1k1[].n2k1.n3k1[].n4k1 ASC,.n4k2.n5k1 DESC,.n4k2 ASC"))
			})
		})
		Context("failure", func() {
			It("should return error on multiple array reference", func() {
				keys := []mongo.Key{
					{Field: "k2[].n1k1[].n2k1.n3k1", Order: 1},
					{Field: "k2[].n1k1[].n2k1.n3k2.n4k1", Order: -1},
					{Field: "k2[].n1k1[].n2k2[]", Order: 1},
				}
				_, err := mongo.GroupAndCombine(keys, false)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("multiple array reference"))
			})
		})
	})
	Describe("generate array expression", func() {
		Context("success", func() {
			It("output data should match with the test data", func() {
				input := "k2[].n1k1[].n2k1.n3k1 ASC INCLUDE MISSING,.n2k1.n3k2.n4k1 DESC,.n2k2 ASC"
				output := "DISTINCT ARRAY (DISTINCT ARRAY FLATTEN_KEYS(`l2Item`.`n2k1`.`n3k1` ASC INCLUDE MISSING,`l2Item`.`n2k1`.`n3k2`.`n4k1` DESC,`l2Item`.`n2k2` ASC) FOR `l2Item` IN `l1Item`.`n1k1` END) FOR `l1Item` IN `k2` END"
				result := mongo.GenerateCouchbaseArrayIndex(input)
				Expect(result).To(Equal(output))
			})
			It("output data should match with the test data", func() {
				input := "k2[].n1k1.n2k1[].n3k1.n4k1[].n5k1 ASC INCLUDE MISSING,.n5k2.n6k1.n7k1 DESC,.n5k3 ASC"
				output := "DISTINCT ARRAY (DISTINCT ARRAY (DISTINCT ARRAY FLATTEN_KEYS(`l3Item`.`n5k1` ASC INCLUDE MISSING,`l3Item`.`n5k2`.`n6k1`.`n7k1` DESC,`l3Item`.`n5k3` ASC) FOR `l3Item` IN `l2Item`.`n3k1`.`n4k1` END) FOR `l2Item` IN `l1Item`.`n1k1`.`n2k1` END) FOR `l1Item` IN `k2` END"
				result := mongo.GenerateCouchbaseArrayIndex(input)
				Expect(result).To(Equal(output))
			})
		})
	})
	Describe("generate partial filter expression", func() {
		Context("success", func() {
			It("output data should match with the test data", func() {
				fieldPath := mongo.IndexFieldPath{}
				fieldPath["k2.n1k1.n2k1.n3k1"] = "k2[].n1k1[].n2k1.n3k1"
				fieldPath["k2.n1k1.n2k1.n3k2.n4k1"] = "k2[].n1k1[].n2k1.n3k2.n4k1"
				fieldPath["k2.n1k1.n2k2"] = "k2[].n1k1[].n2k2"
				output := "ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (`l2Item`.`n2k1`.`n3k1` = 1) END) END"
				fieldExpression, err := mongo.ProcessField("k2.n1k1.n2k1.n3k1", 1, fieldPath)
				Expect(err).To(BeNil())
				Expect(fieldExpression).To(Equal(output))
			})
		})
	})
	Describe("generate partial filter expression", func() {
		Context("success", func() {
			It("output data should match with the test data", func() {
				partialFilter := map[string]interface{}{
					"k1.n1k1": 1,
					"$and": []interface{}{
						map[string]interface{}{
							"k5": 1,
						},
						map[string]interface{}{
							"$or": []interface{}{
								map[string]interface{}{
									"k2.n1k1.n2k1.n3k1": int64(5),
								},
								map[string]interface{}{
									"k2.n1k1.n2k2": float64(10),
								},
							},
						},
					},
				}
				fieldPath := mongo.IndexFieldPath{}
				fieldPath["k2.n1k1.n2k1.n3k1"] = "k2[].n1k1[].n2k1.n3k1"
				fieldPath["k2.n1k1.n2k1.n3k2.n4k1"] = "k2[].n1k1[].n2k1.n3k2.n4k1"
				fieldPath["k2.n1k1.n2k2"] = "k2[].n1k1[].n2k2"

				output := "WHERE (`k1`.`n1k1` = 1 AND (`k5` = 1 AND (ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (`l2Item`.`n2k1`.`n3k1` = 5) END) END OR ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (`l2Item`.`n2k2` = 10) END) END)))"
				result, err := mongo.ConvertMongoToCouchbase(partialFilter, fieldPath)
				Expect(err).To(BeNil())
				if result != output {
					fmt.Println("\n" + result)
					fmt.Println("\n" + output)
				}
				Expect(result).To(Equal(output))
			})
			It("generate partial filter expression with type", func() {
				partialFilter := map[string]interface{}{
					"a": map[string]interface{}{
						"$type": int32(1),
					},
					"b": map[string]interface{}{
						"$type": "string",
					},
				}
				fieldPath := mongo.IndexFieldPath{}
				fieldPath["k2.n1k1.n2k1.n3k1"] = "k2[].n1k1[].n2k1.n3k1"
				fieldPath["k2.n1k1.n2k1.n3k2.n4k1"] = "k2[].n1k1[].n2k1.n3k2.n4k1"
				fieldPath["k2.n1k1.n2k2"] = "k2[].n1k1[].n2k2"

				output := "WHERE (type(`a`) = \"number\" AND type(`b`) = \"string\")"
				result, err := mongo.ConvertMongoToCouchbase(partialFilter, fieldPath)
				Expect(err).To(BeNil())
				if result != output {
					fmt.Println("\n" + result)
					fmt.Println("\n" + output)
				}
				Expect(result).To(Equal(output))
			})
			//It("output data should match with the test data", func() {
			//	input := "k2[].n1k1.n2k1[].n3k1.n4k1[].n5k1 ASC INCLUDE MISSING,.n5k2.n6k1.n7k1 DESC INCLUDE MISSING,.n5k3 ASC INCLUDE MISSING"
			//	output := "DISTINCT ARRAY (DISTINCT ARRAY (DISTINCT ARRAY FLATTEN_KEYS(`l3Item`.`n5k1` ASC INCLUDE MISSING,`l3Item`.`n5k2`.`n6k1`.`n7k1` DESC INCLUDE MISSING,`l3Item`.`n5k3` ASC INCLUDE MISSING) FOR `l3Item` IN `l2Item`.`n3k1`.`n4k1` END) FOR `l2Item` IN `l1Item`.`n1k1`.`n2k1` END) FOR `l1Item` IN `k2` END"
			//	result := mongo.GenerateCouchbaseArrayIndex(input)
			//	Expect(result).To(Equal(output))
			//})
		})
	})

	Describe("Create index query", func() {
		bucket := "bucket1"
		scope := "scope1"
		collection := "collection1"
		Context("success", func() {
			It("output data should match with the test data", func() {
				index := mongo.Index{
					Name: "test",
					Keys: []mongo.Key{
						{Field: "k1.n1k1", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k1", Order: 1},
						{Field: "k3", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k2.n4k1", Order: -1},
						{Field: "k2.n1k1.n2k2", Order: 1},
						{Field: "k4.n1k2.n2k1", Order: 1},
					},
					Sparse: false,
				}
				fieldPath := mongo.IndexFieldPath{}
				fieldPath["k2.n1k1.n2k1.n3k1"] = "k2[].n1k1[].n2k1.n3k1"
				fieldPath["k2.n1k1.n2k1.n3k2.n4k1"] = "k2[].n1k1[].n2k1.n3k2.n4k1"
				fieldPath["k2.n1k1.n2k2"] = "k2[].n1k1[].n2k2"
				arrayExpression := "DISTINCT ARRAY (DISTINCT ARRAY FLATTEN_KEYS(`l2Item`.`n2k1`.`n3k1` ASC,`l2Item`.`n2k1`.`n3k2`.`n4k1` DESC,`l2Item`.`n2k2` ASC) FOR `l2Item` IN `l1Item`.`n1k1` END) FOR `l1Item` IN `k2` END"
				fields := []string{
					"`k1`.`n1k1` ASC INCLUDE MISSING",
					arrayExpression,
					"`k3` ASC",
					"`k4`.`n1k2`.`n2k1` ASC",
				}

				Output := fmt.Sprintf(
					"create index `%s` on `%s`.`%s`.`%s` (%s) ",
					index.Name, bucket, scope, collection, strings.Join(fields, ","))
				query, err := mongo.CreateIndexQuery(bucket, scope, collection, index, fieldPath)
				Expect(err).To(BeNil())
				if query != Output {
					fmt.Println("\n" + query)
					fmt.Println("\n" + Output)
				}
				Expect(query).To(Equal(Output))

			})
			It("output data should match with the test data with partial filter", func() {
				//pfStr := `{"k1.n1k1":1, "$and":[{"k5":1},{"$or": [{"k2.n1k1.n2k1.n3k1": 5}, {"k2.n1k1.n2k2": 10}, {"k2.n1k1.n2k1.n3k2.n4k1" : {"$gte": 100}}]}]}`
				partialFilter := map[string]interface{}{
					"k1.n1k1": 1,
					"$and": []interface{}{
						map[string]interface{}{
							"k5": 1,
						},
						map[string]interface{}{
							"$or": []interface{}{
								map[string]interface{}{
									"k2.n1k1.n2k1.n3k1": int64(5),
								},
								map[string]interface{}{
									"k2.n1k1.n2k2": float64(10),
								},
								map[string]interface{}{
									"k2.n1k1.n2k1.n3k2.n4k1": map[string]interface{}{
										"$gte": 100,
									},
								},
							},
						},
					},
				}

				index := mongo.Index{
					Name:              "test",
					PartialExpression: partialFilter,
					Keys: []mongo.Key{
						{Field: "k1.n1k1", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k1", Order: 1},
						{Field: "k3", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k2.n4k1", Order: -1},
						{Field: "k2.n1k1.n2k2", Order: 1},
						{Field: "k4.n1k2.n2k1", Order: 1},
					},
					Sparse: false,
				}
				fieldPath := mongo.IndexFieldPath{}
				fieldPath["k2.n1k1.n2k1.n3k1"] = "k2[].n1k1[].n2k1.n3k1"
				fieldPath["k2.n1k1.n2k1.n3k2.n4k1"] = "k2[].n1k1[].n2k1.n3k2.n4k1"
				fieldPath["k2.n1k1.n2k2"] = "k2[].n1k1[].n2k2"
				arrayExpression := "DISTINCT ARRAY (DISTINCT ARRAY FLATTEN_KEYS(`l2Item`.`n2k1`.`n3k1` ASC,`l2Item`.`n2k1`.`n3k2`.`n4k1` DESC,`l2Item`.`n2k2` ASC) FOR `l2Item` IN `l1Item`.`n1k1` END) FOR `l1Item` IN `k2` END"
				fields := []string{
					"`k1`.`n1k1` ASC INCLUDE MISSING",
					arrayExpression,
					"`k3` ASC",
					"`k4`.`n1k2`.`n2k1` ASC",
				}

				partialExpression := "WHERE (`k1`.`n1k1` = 1 AND (`k5` = 1 AND (ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (`l2Item`.`n2k1`.`n3k1` = 5) END) END OR ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (`l2Item`.`n2k2` = 10) END) END OR ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (`l2Item`.`n2k1`.`n3k2`.`n4k1` >= 100) END) END)))"
				Output := fmt.Sprintf(
					"create index `%s` on `%s`.`%s`.`%s` (%s) %s",
					index.Name, bucket, scope, collection, strings.Join(fields, ","), partialExpression)
				query, err := mongo.CreateIndexQuery(bucket, scope, collection, index, fieldPath)
				Expect(err).To(BeNil())
				if query != Output {
					fmt.Println("\n" + query)
					fmt.Println(Output)
				}
				Expect(query).To(Equal(Output))
			})
		})
		Context("failure", func() {
			It("output data should match with the test data", func() {
				index := mongo.Index{
					Name: "test",
					Keys: []mongo.Key{
						{Field: "k1.n1k1", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k1", Order: 1},
						{Field: "k3", Order: 1},
						{Field: "k2.n1k1.n2k1.n3k2.n4k1", Order: -1},
						{Field: "k2.n1k1.n2k2", Order: 1},
						{Field: "k4", Order: 1},
					},
					Sparse: false,
				}

				fieldPath := mongo.IndexFieldPath{}
				fieldPath["k2.n1k1.n2k1.n3k1"] = "k2[].n1k1[].n2k1.n3k1"
				fieldPath["k2.n1k1.n2k1.n3k2.n4k1"] = "k2[].n1k1[].n2k1.n3k2.n4k1"
				fieldPath["k2.n1k1.n2k2"] = "k2[].n1k1[].n2k2"
				fieldPath["k4"] = "k4[]"

				_, err := mongo.CreateIndexQuery(bucket, scope, collection, index, fieldPath)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("multiple array reference"))
			})
		})
	})
})
