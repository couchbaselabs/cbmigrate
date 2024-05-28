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
	Parts []DocumentKeyPart
}

type ICBDocumentKey interface {
	Set(key []DocumentKeyPart)
	IsSet() bool
	GetKey() []DocumentKeyPart
	GetNonCompoundPrimaryKeyOnly() string
}

func (k *CBDocumentKey) Set(key []DocumentKeyPart) {
	k.Parts = key
}

func (k *CBDocumentKey) IsSet() bool {
	if len(k.Parts) == 0 {
		return false
	}
	return true
}

func (k *CBDocumentKey) GetKey() []DocumentKeyPart {
	return k.Parts
}

// GetNonCompoundPrimaryKeyOnly value only when the DocumentKeyParts length is 1 and it is a string
func (k *CBDocumentKey) GetNonCompoundPrimaryKeyOnly() string {
	if len(k.Parts) == 1 && k.Parts[0].Kind == DkField {
		return k.Parts[0].Value
	}
	return ""
}

// NewCBDocumentKey returns Singleton DocumentKey as it is used by source (for generating the doc key) and destination
// (for analyzing the index)
func NewCBDocumentKey() ICBDocumentKey {
	return new(CBDocumentKey)
}
