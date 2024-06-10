package mongo_test

import (
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/cbmigrate/internal/mongo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
)

var _ = Describe("index analyzer", func() {
	Describe("navigate path function", func() {
		Context("success", func() {
			It("output data should match with the test data with sparse false", func() {

				jsondata := `{"k1":[null,{},{"n1k1":[10]},{"n1k2":[null,"string",1,{"n2k1":[10]}]}]}`
				indexKeyPath := "k1.n1k2.n2k1"
				var data = map[string]interface{}{}
				err := bson.UnmarshalExtJSON([]byte(jsondata), false, &data)
				if err != nil {
					fmt.Println(err)
				}
				output, found := mongo.NavigatePath(indexKeyPath, data)
				Expect(found).To(BeTrue())
				Expect(output).To(Equal("k1[].n1k2[].n2k1[]"))
			})
			It("output data should match with the test data with sparse false", func() {
				jsondata := `{"k1":{"n1k1": [{"n2k1": {"n3k1": [{"n4k1":1}]}}]}}`
				indexKeyPath := "k1.n1k1.n2k1.n3k1.n4k1"
				var data = map[string]interface{}{}
				err := json.Unmarshal([]byte(jsondata), &data)
				if err != nil {
					fmt.Println(err)
				}
				output, found := mongo.NavigatePath(indexKeyPath, data)
				Expect(found).To(BeTrue())
				Expect(output).To(Equal("k1.n1k1[].n2k1.n3k1[].n4k1"))
			})
		})

	})
})
