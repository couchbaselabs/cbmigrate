package flag

type Flag interface {
	ParseToString() string
	isFlag() bool
	GetName() string
	IsRequired() bool
	IsHidden() bool
	UniqueKey() string
}
