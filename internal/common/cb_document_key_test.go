package common_test

import (
	"github.com/couchbaselabs/cbmigrate/internal/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("couchbase service", func() {
	Describe("test couchbase connection initialization", func() {
		Context("cb document key", func() {
			It("with only primary key", func() {
				docKey := common.NewCBDocumentKey()
				docKey.Set([]common.DocumentKeyPart{{Value: "id", Kind: common.DkField}})
				Expect(docKey.IsSet()).To(Equal(true))
				Expect(docKey.GetKey()).To(Equal([]common.DocumentKeyPart{{Value: "id", Kind: common.DkField}}))
				Expect(docKey.GetNonCompoundPrimaryKeyOnly()).To(Equal("id"))
			})
			It("with string prefix and primary key", func() {
				docKey := common.NewCBDocumentKey()
				docKey.Set([]common.DocumentKeyPart{{Value: "airline", Kind: common.DkString}, {Value: "id", Kind: common.DkField}})
				Expect(docKey.IsSet()).To(Equal(true))
				Expect(docKey.GetKey()).To(Equal([]common.DocumentKeyPart{{Value: "airline", Kind: common.DkString}, {Value: "id", Kind: common.DkField}}))
				Expect(docKey.GetNonCompoundPrimaryKeyOnly()).To(Equal(""))
			})
			It("with not set", func() {
				docKey := common.NewCBDocumentKey()
				docKey.Set([]common.DocumentKeyPart(nil))
				Expect(docKey.IsSet()).To(Equal(false))
				Expect(docKey.GetKey()).To(Equal([]common.DocumentKeyPart(nil)))
				Expect(docKey.GetNonCompoundPrimaryKeyOnly()).To(Equal(""))
			})
			It("with key part zero", func() {
				docKey := common.NewCBDocumentKey()
				docKey.Set([]common.DocumentKeyPart{})
				Expect(docKey.IsSet()).To(Equal(false))
				Expect(docKey.GetKey()).To(Equal([]common.DocumentKeyPart{}))
				Expect(docKey.GetNonCompoundPrimaryKeyOnly()).To(Equal(""))
			})
		})
	})
})
