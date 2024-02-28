package mongo_test

import (
	"fmt"
	"github.com/couchbaselabs/cbmigrate/internal/mongo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
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
	Describe("process field in array filter expression", func() {
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
				partialFilter := bson.D{
					{Key: "k1.n1k1", Value: 1},
					{
						Key: "$and",
						Value: bson.A{
							bson.D{
								{
									Key:   "k5",
									Value: 1,
								},
							},
							bson.D{
								{
									Key: "$or",
									Value: bson.A{
										bson.D{
											{Key: "k2.n1k1.n2k1.n3k1", Value: int64(5)},
										},
										bson.D{
											{Key: "k2.n1k1.n2k2", Value: float64(10)},
										},
									},
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
				partialFilter := bson.D{
					{
						Key: "a", Value: bson.D{
							{
								Key: "$type", Value: int32(1),
							},
						},
					},
					{
						Key: "b", Value: bson.D{
							{
								Key: "$type", Value: "string",
							},
						},
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
			It("generate partial filter expression with type array", func() {
				partialFilter := bson.D{
					{
						Key: "a", Value: bson.D{
							{
								Key: "$type", Value: int32(1),
							},
						},
					},
					{
						Key: "b", Value: bson.D{
							{
								Key: "$type", Value: bson.A{"string", "object", int32(4)},
							},
						},
					},
				}
				fieldPath := mongo.IndexFieldPath{}
				fieldPath["k2.n1k1.n2k1.n3k1"] = "k2[].n1k1[].n2k1.n3k1"
				fieldPath["k2.n1k1.n2k1.n3k2.n4k1"] = "k2[].n1k1[].n2k1.n3k2.n4k1"
				fieldPath["k2.n1k1.n2k2"] = "k2[].n1k1[].n2k2"

				output := "WHERE (type(`a`) = \"number\" AND type(`b`) IN [\"string\",\"object\",\"array\"])"
				result, err := mongo.ConvertMongoToCouchbase(partialFilter, fieldPath)
				Expect(err).To(BeNil())
				if result != output {
					fmt.Println("\n" + result)
					fmt.Println("\n" + output)
				}
				Expect(result).To(Equal(output))
			})
			It("generate partial filter expression with multiple level of conditions", func() {
				partialFilter := bson.D{
					{
						Key: "k5", Value: bson.D{
							{
								Key: "$type", Value: int32(1),
							},
						},
					},
					{
						Key: "k6", Value: 10,
					},
					{
						Key: "$and",
						Value: bson.A{
							bson.D{
								{
									Key: "$or",
									Value: bson.A{
										bson.D{
											{
												Key: "k7", Value: bson.D{
													{Key: "$gte", Value: 10}, {Key: "$exists", Value: true},
												},
											},
											{
												Key: "k8", Value: bson.D{
													{Key: "$lte", Value: 100}, {Key: "$exists", Value: true},
												},
											},
										},
										bson.D{
											{
												Key: "k9", Value: bson.D{
													{Key: "$gte", Value: 10}, {Key: "$exists", Value: true},
												},
											},
											{
												Key: "k10", Value: bson.D{
													{Key: "$lte", Value: 100}, {Key: "$exists", Value: true},
												},
											},
										},
									},
								},
							},
							bson.D{
								{Key: "k11", Value: bson.D{{Key: "$gte", Value: 200}, {Key: "$exists", Value: true}}},
							},
						},
					},
				}
				fieldPath := mongo.IndexFieldPath{}
				fieldPath["k2.n1k1.n2k1.n3k1"] = "k2[].n1k1[].n2k1.n3k1"
				fieldPath["k2.n1k1.n2k1.n3k2.n4k1"] = "k2[].n1k1[].n2k1.n3k2.n4k1"
				fieldPath["k2.n1k1.n2k2"] = "k2[].n1k1[].n2k2"

				output := "WHERE (type(`k5`) = \"number\" AND `k6` = 10 AND (((k7 >= 10 AND k7 IS NOT \"NULL\" AND k8 <= 100 AND k8 IS NOT \"NULL\") OR (k9 >= 10 AND k9 IS NOT \"NULL\" AND k10 <= 100 AND k10 IS NOT \"NULL\")) AND k11 >= 200 AND k11 IS NOT \"NULL\"))"
				result, err := mongo.ConvertMongoToCouchbase(partialFilter, fieldPath)
				Expect(err).To(BeNil())
				if result != output {
					fmt.Println("\n" + result)
					fmt.Println("\n" + output)
				}
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
					"create index `%s` on `%s`.`%s`.`%s` (%s)  USING GSI WITH {\"defer_build\":true}",
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

				partialFilter := bson.D{
					{Key: "k1.n1k1", Value: 1},
					{
						Key: "$and",
						Value: bson.A{
							bson.D{
								{Key: "k5", Value: 1},
							},
							bson.D{
								{
									Key: "$or",
									Value: bson.A{
										bson.D{
											{Key: "k2.n1k1.n2k1.n3k1", Value: int64(5)},
										},
										bson.D{
											{Key: "k2.n1k1.n2k2", Value: float64(10)},
										},
										bson.D{
											{
												Key: "k2.n1k1.n2k1.n3k2.n4k1",
												Value: bson.D{
													{Key: "$gte", Value: 100},
												},
											},
										},
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
					"create index `%s` on `%s`.`%s`.`%s` (%s) %s USING GSI WITH {\"defer_build\":true}",
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

	Describe("generate array filter expression", func() {
		Context("success", func() {
			It("case 1 type filter true and tailing field not array", func() {
				key := "k2[].n1k1[].n2k1.n3k1"
				condition := " = \"string\""
				output := mongo.GenerateArrayFilterExpression(key, true, condition)
				Expect(output).To(Equal("ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (type(`l2Item`.`n2k1`.`n3k1`)  = \"string\") END) END"))
			})
			It("case 2 type filter true and tailing field array but type filter is string", func() {
				key := "k2[].n1k1[].n2k1.n3k1[]"
				condition := " = \"string\""
				output := mongo.GenerateArrayFilterExpression(key, true, condition)
				Expect(output).To(Equal("ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (type(`l2Item`.`n2k1`.`n3k1`)  = \"string\") END) END"))
			})
			It("case 3 type filter true and tailing field not array 2", func() {
				key := "k2[].n1k1[].n2k1.n3k1[]"
				condition := " = \"array\""
				output := mongo.GenerateArrayFilterExpression(key, true, condition)
				Expect(output).To(Equal("ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (type(`l2Item`.`n2k1`.`n3k1`)  = \"array\") END) END"))
			})
			It("case 4 type filter true and field  array", func() {
				key := "k2[]"
				condition := " = \"array\""
				output := mongo.GenerateArrayFilterExpression(key, true, condition)
				Expect(output).To(Equal("type(`k2`)  = \"array\""))
			})
			It("case 5 type filter false and field  array", func() {
				key := "k2[].n1k1[].n2k1.n3k1[]"
				condition := " = \"val\""
				output := mongo.GenerateArrayFilterExpression(key, false, condition)
				Expect(output).To(Equal("ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (ANY `l3Item` IN `l2Item`.`n2k1`.`n3k1` SATISFIES (`l3Item`  = \"val\") END) END) END"))
			})
			It("case 6 type filter false and field not array", func() {
				key := "k2[].n1k1[].n2k1.n3k1"
				condition := " = \"val\""
				output := mongo.GenerateArrayFilterExpression(key, false, condition)
				Expect(output).To(Equal("ANY `l1Item` IN `k2` SATISFIES (ANY `l2Item` IN `l1Item`.`n1k1` SATISFIES (`l2Item`.`n2k1`.`n3k1`  = \"val\") END) END"))
			})
		})
	})
})
