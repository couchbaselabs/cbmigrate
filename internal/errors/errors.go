package errors

type NotSupportedError interface {
	IsNotSupportedError() bool
	Error() string
}

type MongoNotSupportedError struct {
	Message string
}

func (e *MongoNotSupportedError) IsNotSupportedError() bool {
	return true
}

func (e *MongoNotSupportedError) Error() string {
	return e.Message
}

func NewMongoNotSupportedError(message string) error {
	return &MongoNotSupportedError{Message: message}
}
