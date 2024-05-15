package common

type DocumentKind string

const (
	DkString DocumentKind = "string"
	DkUuid   DocumentKind = "UUID"
	DkField  DocumentKind = "field"
)

type DocumentKeyPart struct {
	Value string
	Kind  DocumentKind // string | field | UUID
}

// CBDocumentKey can be generated with string and field or a composite filed or uuid using generator syntex
type CBDocumentKey struct {
	Key []DocumentKeyPart
}

type ICBDocumentKey interface {
	Set(key []DocumentKeyPart)
	IsSet(key []DocumentKeyPart) bool
	GetKey() []DocumentKeyPart
	GetPrimaryKeyOnly() string
}

func (k *CBDocumentKey) Set(key []DocumentKeyPart) {
	k.Key = key
}

func (k *CBDocumentKey) IsSet(key []DocumentKeyPart) bool {
	if len(k.Key) == 0 {
		return false
	}
	return true
}

func (k *CBDocumentKey) GetKey() []DocumentKeyPart {
	return k.Key
}

// GetPrimaryKeyOnly value only when the DocumentKeyParts length is 1 and it is a string
func (k *CBDocumentKey) GetPrimaryKeyOnly() string {
	if len(k.Key) == 1 && k.Key[0].Kind == DkString {
		return k.Key[0].Value
	}
	return ""
}

// NewCBDocumentKey returns Singleton DocumentKey as it is used by source (for generating the doc key) and destination
// (for analyzing the index)
func NewCBDocumentKey() ICBDocumentKey {
	return new(CBDocumentKey)
}
