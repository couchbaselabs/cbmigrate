package index

type Index struct {
	Name              string
	Keys              []Key
	PartialExpression map[string]interface{}
	Unique            bool
	Sparse            bool
	NotSupported      bool
}
type Key struct {
	Field string
	Order int
}

// IndexFieldPath is used to have array representation for a particular path
// example: k1.n1k1.n2k1 is path for field n2k1 in a document. n1k1 is an array, and it is represented as k1.n1k1[].n2k1.
type IndexFieldPath map[string]string

func (i IndexFieldPath) Get(key string) string {
	if i == nil {
		return key
	}
	v := i[key]
	if v == "" {
		return key
	}
	return v
}
